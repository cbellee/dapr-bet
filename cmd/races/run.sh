go build -o bets-service . && dapr run --app-id bets --app-protocol http --app-port 8004 --components-path ./components --log-level info ./server
