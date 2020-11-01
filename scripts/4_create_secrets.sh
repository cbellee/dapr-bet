kubectl create secret generic cosmos-db-secret --from-env-file=../.env
kubectl create secret generic cosmos-db-secret --from-env-file=../.env --namespace=dapr-monitoring
kubectl create secret docker-registry registry-secret --docker-server=daprbet891237.azurecr.io --docker-username=daprbet891237 --docker-password=tqE90vTLfmatvZPRd5IZ/di8epwOkF3e --docker-email=cbellee@microsoft.com

# list & decode secrets
kubectl get secrets/cosmos-db-secret --template={{.data.cosmosDbConnectionString}} | base64 -d
kubectl get secrets/cosmos-db-secret --template={{.data.cosmosDbMasterKey}} | base64 -d
kubectl get secrets/cosmos-db-secret --template={{.data.cosmosDbUrl}} | base64 -d
kubectl get secrets/cosmos-db-secret --template={{.data.aiInstrumentationKey}} | base64 -d
kubectl get secrets/cosmos-db-secret --template={{.data.aiApiKey}} | base64 -d
kubectl get secrets/cosmos-db-secret --template={{.data.sbConnectionString}} | base64 -d

# secrets