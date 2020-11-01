helm repo add stable https://kubernetes-charts.storage.googleapis.com/
helm install nginx stable/nginx-ingress -f ./manifests/nginx_dapr_annotations.yml -n default

kubectl apply -f ./manifests/nginx_ingress.yml