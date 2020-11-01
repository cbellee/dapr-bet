VERSION := 0.1.0
ENVIRONMENT := dev
TAG := ${ENVIRONMENT}-${VERSION}
BUILD_INFO := "dapr bet demo"
ACR_LOGIN_NAME := daprbet891237
ACR_URI := daprbet891237.azurecr.io
RESULTS_SERVICE_PORT := 8003
BETS_SERVICE_PORT := 8004
PUNTERS_SERVICE_PORT := 8002
RACES_SERVICE_PORT := 8005

#### reference existing shell env vars #####
# ${cosmosDbConnectionString}
# ${cosmosDbMasterKey}
# ${aiInstrumentationKey}
# ${aiApiKey}
# ${acrAdminPassword}
# ${sbConnectionString}

build_results:
	docker build -t ${ACR_URI}/results:${TAG} --build-arg SERVICE_NAME="results" --build-arg SERVICE_PORT=${RESULTS_SERVICE_PORT} --build-arg BUILD_INFO=${BUILD_INFO} --build-arg VERSION=${VERSION} -f Dockerfile .
	docker image prune -f

build_bets:
	docker build -t ${ACR_URI}/bets:${TAG} --build-arg SERVICE_NAME="bets" --build-arg SERVICE_PORT=${BETS_SERVICE_PORT} --build-arg BUILD_INFO=${BUILD_INFO} --build-arg VERSION=${VERSION} -f Dockerfile .
	docker image prune -f

build_punters:
	docker build -t ${ACR_URI}/punters:${TAG} --build-arg SERVICE_NAME="punters" --build-arg SERVICE_PORT=${PUNTERS_SERVICE_PORT} --build-arg BUILD_INFO=${BUILD_INFO} --build-arg VERSION=${VERSION} -f Dockerfile .
	docker image prune -f

build_races:
	docker build -t ${ACR_URI}/races:${TAG} --build-arg SERVICE_NAME="races" --build-arg SERVICE_PORT=${RACES_SERVICE_PORT} --build-arg BUILD_INFO=${BUILD_INFO} --build-arg VERSION=${VERSION} -f Dockerfile .
	docker image prune -f


build: build_results build_bets build_punters build_races

push_results:
	docker login ${ACR_URI} -u ${ACR_LOGIN_NAME} -p ${acrAdminPassword}
	docker push ${ACR_URI}/results:${TAG}

push_bets:
	docker login ${ACR_URI} -u ${ACR_LOGIN_NAME} -p ${acrAdminPassword}
	docker push ${ACR_URI}/bets:${TAG}

push_punters:
	docker login ${ACR_URI} -u ${ACR_LOGIN_NAME} -p ${acrAdminPassword}
	docker push ${ACR_URI}/punters:${TAG}

push_races:
	docker login ${ACR_URI} -u ${ACR_LOGIN_NAME} -p ${acrAdminPassword}
	docker push ${ACR_URI}/races:${TAG}

push: push_results push_bets push_punters push_races

deploy_results:
	kubectl delete deploy results
	kubectl apply -f ./manifests/app_results.yml

deploy_bets:
	kubectl delete deploy bets
	kubectl apply -f ./manifests/app_bets.yml

deploy_punters:
	kubectl delete deploy punters
	kubectl apply -f ./manifests/app_punters.yml

deploy_races:
	kubectl delete deploy races
	kubectl apply -f ./manifests/app_races.yml

deploy: deploy_results deploy_bets deploy_punters deploy_races

logs_results:
	kubectl logs $$(kubectl get pods --selector=app=results --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') results -f

logs_bets:
	kubectl logs $$(kubectl get pods --selector=app=bets --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') bets -f

logs_punters:
	kubectl logs $$(kubectl get pods --selector=app=punters --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') punters -f

logs_races:
	kubectl logs $$(kubectl get pods --selector=app=races --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') races -f

logs_results_dapr:
	kubectl logs $$(kubectl get pods --selector=app=results --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') daprd -f

logs_bets_dapr:
	kubectl logs $$(kubectl get pods --selector=app=bets --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') daprd -f

logs_punters_dapr:
	kubectl logs $$(kubectl get pods --selector=app=punters --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') daprd -f

logs_races_dapr:
	kubectl logs $$(kubectl get pods --selector=app=races --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}') daprd -f

deploy_dapr_components:
	# deploy AAD pod identity 
	kubectl apply -f ./manifests/azure_identity_config.yml
	kubectl apply -f ./manifests/dapr_azure_keyvault.yml
	
	# add secrets to key vault
	#az keyvault secret set --name cosmosDbConnectionString --vault-name dapr-bet-kv --value ${cosmosDbConnectionString}
	#az keyvault secret set --name cosmosDbMasterKey --vault-name dapr-bet-kv --value ${cosmosDbMasterKey}
	#az keyvault secret set --name cosmosDbUrl --vault-name dapr-bet-kv --value ${cosmosDbUrl}
	#az keyvault secret set --name aiInstrumentationKey --vault-name dapr-bet-kv --value ${aiInstrumentationKey}
	#az keyvault secret set --name aiApiKey --vault-name dapr-bet-kv --value ${aiApiKey}
	#az keyvault secret set --name signalrConnectionString --vault-name dapr-bet-kv --value ${signalrConnectionString}
	#az keyvault secret set --name sbConnectionString --vault-name dapr-bet-kv --value ${sbConnectionString}

	# apply ai forwarder manifests
	kubectl apply -f ./manifests/dapr_ai_forwarder.yml
	kubectl apply -f ./manifests/dapr_ai_native_exporter.yml
	kubectl apply -f ./manifests/tracing.yml

	# apply dapr components
	kubectl apply -f ./manifests/dapr_cosmosdb_binding_bets.yml
	kubectl apply -f ./manifests/dapr_cosmosdb_binding_punters.yml
	kubectl apply -f ./manifests/dapr_cosmosdb_binding_results.yml
	kubectl apply -f ./manifests/dapr_cosmosdb_statestore_bets.yml
	kubectl apply -f ./manifests/dapr_cosmosdb_statestore_punters.yml
	kubectl apply -f ./manifests/dapr_servicebus_pubsub.yml
	kubectl apply -f ./manifests/dapr_signalr_binding.yml
