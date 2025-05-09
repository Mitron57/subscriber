# subscriber
Simple gRPC publisher-subscriber service

## Features
- Onion architecture
- Graceful shutdown on SIGINT, SIGQUIT, SIGTERM
- Dependency injection
- Logging with [zap](https://github.com/uber-go/zap)
- Configuration via environment

## API
Full gRPC API you can see in [schema](api/schema.proto).
- Subscribe method holds connection until it's cancelled by client. All subscribers will receive the message sent via Publish method.
- Publish method is a classic request. If there's no such subject, corresponding error will be returned (error code 5).

## Build
Service requires 2 environment variables: `SUBSCRIBER_HOST` and `SUBSCRIBER_PORT`, which can be passed via .env file (see [example](.env.example)). You can build this service via make. There are 3 make targets: proto, docker, server.

### Proto
The target simply generates gRPC code for this service in [internal/handlers/grpc/proto](internal/handlers/grpc/proto) and requires `protoc` with go-grpc extension installed in your environment. Use this target only when [protobuf](api/schema.proto) has been modified.

##### Run: `make proto`

### Docker
The target builds subscriber docker image.

##### Build: `make docker`
##### Run: `docker run -e SUBSCRIBER_HOST=0.0.0.0 -e SUBSCRIBER_PORT=50051 -p 127.0.0.1:50051:50051 -dt subscriber:latest`

### Server
The target builds server executable natively and stores executable in `build` directory.

##### Build: `make server`
##### Run: `./build/subscriber`