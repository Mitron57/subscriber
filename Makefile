proto: 
	protoc --go_out=. --go-grpc_out=. api/scheme.proto
docker:
	docker build -t subscriber .
server:
	go build -o ./build/ ./cmd/subscriber