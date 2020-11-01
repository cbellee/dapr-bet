helm repo add dapr https://daprio.azurecr.io/helm/v1/repo
helm repo update
kubectl create namespace dapr-system
helm install dapr dapr/dapr --namespace dapr-system --set dapr_operator.loglevel=error --set dapr_placement.loglevel=error --set dapr_sidecar_injector.loglevel=error

# NOTE - need to add 'watch' permission to 'subscription' for clusterrole 'dapr-operator-admin'
# k edit clusterrole dapr-operator-admin 
# in vi, add '- watch' to 'subscription'
