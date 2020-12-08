go build -o results-service . && dapr run --app-id results --app-protocol http --app-port 8003 --components-path ./components --log-level info ./server
