package grpc

import (
	"context"
	"errors"

	"github.com/mitron57/subpub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"subscriber/internal/domain/dto"
	"subscriber/internal/domain/interfaces"
	"subscriber/internal/handlers/grpc/proto"
)

type Handler struct {
	proto.UnimplementedPubSubServer
	bus    interfaces.PubSub
	logger *zap.Logger
}

func NewPubSubServer(bus interfaces.PubSub, logger *zap.Logger) proto.PubSubServer {
	return &Handler{bus: bus, logger: logger}
}

func (i *Handler) makeHandler(stream grpc.ServerStreamingServer[proto.Event]) subpub.MessageHandler {
	return func(msg any) {
		event := proto.Event{Data: msg.(string)}
		if err := stream.Send(&event); err != nil {
			i.logger.Error("failed to send event", zap.Error(err))
		}
	}
}

// transmuteError reinterprets application-layer errors into gRPC errors
func transmuteError(err error) error {
	if err == nil {
		return nil //basically what status.Error(codes.OK, ...) does
	}
	if errors.Is(err, subpub.ErrNoSuchSubject) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, subpub.ErrClosedBus) {
		return status.Error(codes.Aborted, err.Error())
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func (i *Handler) Subscribe(req *proto.SubscribeRequest, stream grpc.ServerStreamingServer[proto.Event]) error {
	subscription := dto.Subscription{
		Topic:   req.Key,
		Handler: i.makeHandler(stream),
	}
	sub, err := i.bus.Subscribe(stream.Context(), subscription)
	if err != nil {
		return transmuteError(err)
	}
	<-stream.Context().Done()
	sub.Unsubscribe()
	return transmuteError(err)
}

func (i *Handler) Publish(ctx context.Context, req *proto.PublishRequest) (*emptypb.Empty, error) {
	message := dto.Message{
		Topic: req.Key,
		Data:  req.Data,
	}
	err := i.bus.Publish(ctx, message)
	return &emptypb.Empty{}, transmuteError(err)
}
