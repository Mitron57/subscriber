package grpc

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Interceptor struct {
	logger    *zap.Logger
	requestId atomic.Uint64
}

func NewInterceptor(logger *zap.Logger) *Interceptor {
	return &Interceptor{logger: logger}
}

func (i *Interceptor) RequestIdUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id := i.requestId.Add(1)
		ctx = context.WithValue(ctx, "request-id", id)
		return handler(ctx, req)
	}
}

func (i *Interceptor) LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id := ctx.Value("request-id").(uint64)
		i.logger.Info("New RPC request", zap.Uint64("request-id", id), zap.String("method", info.FullMethod))

		result, err := handler(ctx, req)

		statusCode := status.Code(err).String()
		i.logger.Info(
			"Request handled",
			zap.Uint64("request-id", id),
			zap.String("method", info.FullMethod),
			zap.String("status-code", statusCode),
		)
		return result, err
	}
}

type Streamer struct {
	grpc.ServerStream
	ctx context.Context
}

func (i *Interceptor) RequestIdStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id := i.requestId.Add(1)
		ctx := context.WithValue(ss.Context(), "request-id", id)

		wrapped := Streamer{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, &wrapped)
	}
}

func (i *Interceptor) LoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.(*Streamer).ctx
		id := ctx.Value("request-id").(uint64)

		i.logger.Info("New Stream-RPC request", zap.Uint64("request-id", id), zap.String("method", info.FullMethod))

		err := handler(srv, ss)
		statusCode := status.Code(err).String()

		i.logger.Info(
			"Request handled",
			zap.Uint64("request-id", id),
			zap.String("method", info.FullMethod),
			zap.String("status-code", statusCode),
		)
		return err
	}
}
