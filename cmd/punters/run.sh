go build -o punters-service . && dapr run --app-id punters --app-protocol http --app-port 8002 --components-path ./components --log-level info ./server
