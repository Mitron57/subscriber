proto: 
	protoc --go_out=. --go-grpc_out=. api/schema.proto
docker:
	docker build -t subscriber .
server:
	go build -o ./build/ ./cmd/subscriber

test-integration:
	go test -v -tags=integration ./tests/integration/...
