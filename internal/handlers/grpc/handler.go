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
	return &Handler{
		bus:    bus,
		logger: logger,
	}
}

func (h *Handler) makeHandler(stream grpc.ServerStreamingServer[proto.Event]) subpub.MessageHandler {
	return func(msg any) {
		event := proto.Event{
			Data: msg.(string),
		}
		if err := stream.Send(&event); err != nil {
			h.logger.Error("failed to send event", zap.Error(err))
		}
	}
}

func (h *Handler) Subscribe(req *proto.SubscribeRequest, stream grpc.ServerStreamingServer[proto.Event]) error {
	subscription := dto.Subscription{
		Topic:   req.Key,
		Handler: h.makeHandler(stream),
	}

	sub, err := h.bus.Subscribe(stream.Context(), subscription)

	if errors.Is(err, subpub.ErrClosedBus) {
		return status.Error(codes.Aborted, err.Error())
	}
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	<-stream.Context().Done()

	sub.Unsubscribe()

	return nil
}

func (h *Handler) Publish(ctx context.Context, req *proto.PublishRequest) (*emptypb.Empty, error) {
	message := dto.Message{
		Topic: req.Key,
		Data:  req.Data,
	}

	err := h.bus.Publish(ctx, message)
	if errors.Is(err, subpub.ErrClosedBus) {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	if errors.Is(err, subpub.ErrNoSuchSubject) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
