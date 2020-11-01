# export env vars defined in /.env file
export $(xargs < ../.env)

SUBSCRIPTION_ID=$(az account show --query id -o tsv)
LOCATION=australiaeast
APP_NAME=dapr-bet
RG_NAME=${APP_NAME}
ACR_NAME=daprbet891237
AKS_NAME=dapr-bet-cluster
AKS_VERSION=1.18.8
AKS_VM_SKU=Standard_F2s_v2
COSMOS_DB_ACCOUNT_NAME=dapr-bet-account
SERVICE_BUS_NAMESPACE=dapr-bet-sbus
KV_NAME=${APP_NAME}-kv
DB_DATA='[{"name":"bets","partitionKey":"/raceid"},{"name":"punters","partitionKey":"/email"},{"name":"results","partitionKey":"/raceid"}]'
TOPICS=(results payments)

# create resource group
az group create -n $RG_NAME -l $LOCATION

# create cosmosdb account
az cosmosdb create --resource-group $RG_NAME --name $COSMOS_DB_ACCOUNT_NAME --default-consistency-level Strong

# create cosmosdb databases
for row in $(echo "${DB_DATA}" | jq -r '.[] | @base64'); do
    _jq() {echo ${row} | base64 --decode | jq -r ${1}}
    
    az cosmosdb sql database create --resource-group $RG_NAME \
        --account-name $COSMOS_DB_ACCOUNT_NAME \
        --name $(_jq '.name')

    az cosmosdb sql container create --resource-group $RG_NAME \
        --account-name $COSMOS_DB_ACCOUNT_NAME \
        --database-name $(_jq '.name') \
        --name default \
        --partition-key-path $(_jq '.partitionKey') 
done

# create service bus namespace
az servicebus namespace create --resource-group $RG_NAME --name $SERVICE_BUS_NAMESPACE --location $LOCATION

# create servicebus topics & subscriptions
for topic in "${TOPICS[@]}"; do 
    az servicebus topic create --resource-group $RG_NAME \
        --namespace-name $SERVICE_BUS_NAMESPACE \
        --name "$topic"

    az servicebus topic subscription create --resource-group $RG_NAME \
        --namespace-name $SERVICE_BUS_NAMESPACE \
        --topic-name "$topic" \
        --name "$topic"  
done

# get service bus info
az servicebus namespace authorization-rule keys list --resource-group $RG_NAME \
    --namespace-name $SERVICE_BUS_NAMESPACE \
    --name RootManageSharedAccessKey \
    --query primaryConnectionString \
    --output tsv

# create log analytics workspace
WS_ID=$(az monitor log-analytics workspace create --resource-group $RG_NAME --workspace-name $WS_NAME --query id -o tsv)

# create application insights
az monitor app-insights component create --resource-group $RG_NAME --location $LOCATION --app $APP_NAME

# create ACR
ACR=$(az acr create --resource-group $RG_NAME --name $ACR_NAME --sku Standard)

# export as env vars for './scripts/4_create_secrets' & makeFile
export ACR_LOGIN_NAME=$(echo $ACR | jq '.name' -r)
export ACR_URI=$(echo $ACR | jq '.loginServer' -r)

# create keyvault
az keyvault create --location $LOCATION --name $KV_NAME --resource-group $RG_NAME

# create vnet
az network vnet create --resource-group $RG_NAME \
	--name "${AKS_NAME}-vnet" \
	--address-prefix 10.0.0.0/16 \
	--subnet-name "${AKS_NAME}-subnet" \
	--subnet-prefix 10.0.0.0/24

# create ACI subnet for virtual kubelet scaling
az network vnet subnet create --resource-group $RG_NAME \
	--vnet-name "${AKS_NAME}-vnet" \
	--name "aci-subnet" \
	--address-prefixes 10.0.2.0/24

SUBNET_ID=$(az network vnet subnet show --resource-group $RG_NAME --vnet-name "${AKS_NAME}-vnet" --name "${AKS_NAME}-subnet" --query id -o tsv)

# get 'aks-admin-group' objectId
ADMIN_GRP_OID=$(az ad group list --display-name aks-admin-group --query [].objectId -o tsv)

# create AKS cluster
az aks create --resource-group $RG_NAME \
	--name $AKS_NAME \
	--kubernetes-version $AKS_VERSION \
	--ssh-key-value "$(cat ~/.ssh/id_rsa.pub)" \
	--node-vm-size $AKS_VM_SKU \
	--node-count 2 \
	--enable-cluster-autoscaler \
	--service-cidr 10.0.1.0/24 \
	--dns-service-ip 10.0.1.10 \
	--min-count 1 \
	--max-count 5 \
	--network-plugin azure \
	--vnet-subnet-id $SUBNET_ID \
	--enable-managed-identity \
	--attach-acr $ACR_NAME \
	--enable-aad \
	--aad-admin-group-object-ids $ADMIN_GRP_OID

# enable virtual nodes addon
az aks enable-addons \
    --resource-group $RG_NAME \
    --name $AKS_NAME \
    --addons virtual-node \
    --subnet-name aci-subnet

# enable container insights for AKS cluster
az aks enable-addons -a monitoring -n $AKS_NAME -g $RG_NAME --workspace-resource-id $WS_ID

# get AKS managed identity
AKS_MID_CLIENT_ID=$(az aks show --resource-group $RG_NAME --name $AKS_NAME --query identityProfile.kubeletidentity.clientId -o tsv)
AKS_NODE_RG_NAME=$(az aks show --resource-group $RG_NAME --name $AKS_NAME --query nodeResourceGroup -o tsv)

# assign roles to managed identity
az role assignment create \
    --role "Reader" \
	--assignee $AKS_MID_CLIENT_ID \
	--scope "/subscriptions/${SUBSCRIPTION_ID}/resourcegroups/${RG_NAME}"

az role assignment create \
    --role "Managed Identity Operator"  \
	--assignee $AKS_MID_CLIENT_ID \
	--scope "/subscriptions/${SUBSCRIPTION_ID}/resourcegroups/${RG_NAME}"

az role assignment create \
    --role "Virtual Machine Contributor" \
	--assignee $AKS_MID_CLIENT_ID \
	--scope "/subscriptions/${SUBSCRIPTION_ID}/resourcegroups/${RG_NAME}"

# set key vault policy 
az keyvault set-policy --name $KV_NAME --spn $AKS_MID_CLIENT_ID --secret-permissions get list

# deploy aad pod identity using Helm
helm repo add aad-pod-identity https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts
helm install aad-pod-identity aad-pod-identity/aad-pod-identity

# enable pod identity
#kubectl apply -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/deployment-rbac.yaml

# For AKS clusters, deploy the MIC and AKS add-on exception by running
#kubectl apply -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/mic-exception.yaml