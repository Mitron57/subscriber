package grpc

//go:generate mockery --name=MockServerStream --output=./mocks --with-expecter

import (
	"google.golang.org/grpc"

	"subscriber/internal/handlers/grpc/proto"
)

// MockServerStream is an interface that wraps the grpc.ServerStreamingServer interface
type MockServerStream interface {
	grpc.ServerStreamingServer[proto.Event]
}
