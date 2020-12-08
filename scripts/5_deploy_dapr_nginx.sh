helm repo add stable https://kubernetes-charts.storage.googleapis.com/
helm install nginx stable/nginx-ingress -f ./manifests/dapr.nginx.annotations.yml -n default

kubectl apply -f ./manifests/nginx.ingress.yml