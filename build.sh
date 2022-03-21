rm -rf ./data && mkdir data
go build -o ./build/miner ./cmd/miner_client.go
go build -o ./build/user ./cmd/user_client.go