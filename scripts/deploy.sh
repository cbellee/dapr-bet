LOCATION='australiaeast'
ENVIRONMENT='dev'
SSH_KEY=$(cat ~/.ssh/id_rsa.pub)
ADMIN_GROUP_OBJECT_ID="f6a900e2-df11-43e7-ba3e-22be99d3cede"
RG_NAME="aks-dapr-bet-rg"
VERSION='0.1.0'
TAG=${ENVIRONMENT}-${VERSION}
SERVICES=(bets results punters races)
PUNTERS_SERVICE_PORT=8002
RESULTS_SERVICE_PORT=8003
BETS_SERVICE_PORT=8004
RACES_SERVICE_PORT=8005
AZURE_TENANT_ID=$(az account show --query 'tenantId' -o tsv)
APP_NAME="dapr-application"

# Create app registration & service principal
APP_ID=$(az ad app create --display-name "${APP_NAME}"  | jq -r .appId)

az ad app credential reset \
  --id "${APP_ID}" \
  --years 2

SERVICE_PRINCIPAL_ID=$(az ad sp create \
  --id "${APP_ID}" \
  | jq -r .id)

echo "Service Principal ID: ${SERVICE_PRINCIPAL_ID}"

# az feature register --namespace "Microsoft.ContainerService" --name "AKS-KedaPreview"

az group create --location $LOCATION --name $RG_NAME

az deployment group create \
    --resource-group $RG_NAME \
    --name aks-dapr-bet-$ENVIRONMENT-deployment \
    --template-file ../main.bicep \
    --parameters @main.parameters.json \
    --parameters environment=$ENVIRONMENT \
    --parameters sshPublicKey="$SSH_KEY" \
    --parameters adminGroupObjectID=$ADMIN_GROUP_OBJECT_ID

ACR_NAME=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.acrName.value' -o tsv)
CLUSTER_NAME=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.aksClusterName.value' -o tsv)
AZURE_CLIENT_ID=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.midPrincipalId.value' -o tsv)
VAULT_NAME=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.keyVaultName.value' -o tsv)
COSMOS_DB_KEY=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.cosmosDbKey.value' -o tsv)
COSMOS_DB_URL=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.cosmosDbUrl.value' -o tsv)
# SBUS_CONNECTION_STRING=$(az deployment group show --resource-group $RG_NAME --name aks-dapr-bet-$ENVIRONMENT-deployment --query 'properties.outputs.serviceBusConnectionString.value' -o tsv)

az aks get-credentials -g $RG_NAME -n $CLUSTER_NAME --admin --overwrite-existing --context "aks-dapr-bet-$ENVIRONMENT"

cd ..

az acr build --registry $ACR_NAME \
    --image dapr-bet/bets:${TAG} \
    --build-arg SERVICE_NAME="bets" \
    --build-arg SERVICE_PORT=${BETS_SERVICE_PORT} \
    --build-arg VERSION=${VERSION} \
    -f Dockerfile .

az acr build --registry $ACR_NAME \
    --image dapr-bet/results:${TAG} \
    --build-arg SERVICE_NAME="results" \
    --build-arg SERVICE_PORT=${RESULTS_SERVICE_PORT} \
    --build-arg VERSION=${VERSION} \
    -f Dockerfile .

az acr build --registry $ACR_NAME \
    --image dapr-bet/punters:${TAG} \
    --build-arg SERVICE_NAME="punters" \
    --build-arg SERVICE_PORT=${PUNTERS_SERVICE_PORT} \
    --build-arg VERSION=${VERSION} \
    -f Dockerfile .

az acr build --registry $ACR_NAME \
    --image dapr-bet/races:${TAG} \
    --build-arg SERVICE_NAME="races" \
    --build-arg SERVICE_PORT=${RACES_SERVICE_PORT} \
    --build-arg VERSION=${VERSION} \
    -f Dockerfile .

cd ./scripts

az aks update -g $RG_NAME -n $CLUSTER_NAME --enable-keda

# apply ai forwarder manifests
kubectl apply -f ../manifests/appinsights.forwarder.yml
kubectl apply -f ./components/dapr-bet.secretstore.yml
# kubectl apply -f ./manifests/dapr-bet.exporter.appinsights.yml
# kubectl apply -f ./components/dapr-bet.tracing.yml

sed "s|<VAULT_NAME>|$VAULT_NAME|g;s|<AZURE_CLIENT_ID>|$AZURE_CLIENT_ID|g;s|<AZURE_TENANT_ID>|$AZURE_TENANT_ID|g" ../components/dapr-bet.secretstore.yml | kubectl apply -f -
# sed "s|<VAULT_NAME>|$VAULT_NAME|g" ../components/dapr-bet.secretstore.yml | kubectl apply -f -

kubectl create secret generic cosmos-secrets --from-literal=cosmos-db-master-key=$COSMOS_DB_KEY --from-literal=cosmos-db-url=$COSMOS_DB_URL

helm repo add nginx-stable https://helm.nginx.com/stable
helm repo add stable https://charts.helm.sh/stable
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.13.1 \
  --set installCRDs=true

helm install nginx-ingress nginx-stable/nginx-ingress --set rbac.create=true

kubectl apply -f ../manifests/keda.results.scaler.yml
kubectl apply -f ../manifests/lets.encrypt.yml
kubectl apply -f ../manifests/nginx.ingress.yml

# apply dapr components
kubectl apply -f ../components/dapr-bet.binding.bets.yml
kubectl apply -f ../components/dapr-bet.binding.punters.yml
kubectl apply -f ../components/dapr-bet.binding.results.yml
kubectl apply -f ../components/dapr-bet.statestore.bets.yml
kubectl apply -f ../components/dapr-bet.statestore.punters.yml
kubectl apply -f ../components/dapr-bet.pubsub.servicebus.yml
#kubectl apply -f ./components/dapr-bet.binding.signalr.yml

# patch & apply service manifests
for SERVICE in "${SERVICES[@]}"
do
    sed "s|<IMAGE_NAME>|$ACR_NAME.azurecr.io/dapr-bet/$SERVICE:${TAG}|g" ../manifests/deploy.$SERVICE.yml | kubectl apply -f -
done

