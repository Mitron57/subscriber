proto: 
	protoc --go_out=. --go-grpc_out=. api/scheme.proto
all:
	go build ./cmd/subscriber