package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mitron57/subpub"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"subscriber/config"
	handlers "subscriber/internal/handlers/grpc"
	"subscriber/internal/handlers/grpc/proto"
	"subscriber/internal/services"
)

type Server struct {
	grpcHandler proto.PubSubServer
	server      *grpc.Server
	bus         subpub.SubPub
	config      *config.Config
	logger      *zap.Logger
}

func NewServer(config *config.Config, logger *zap.Logger) *Server {
	return &Server{
		config: config,
		logger: logger,
	}
}

func (s *Server) gracefulStop() error {
	s.logger.Info("Shutting down server")
	s.server.GracefulStop()
	return s.bus.Close(context.Background())
}

func (s *Server) Init() {
	s.bus = subpub.NewPubSub()
	service := services.NewPubSub(s.bus)

	handler := handlers.NewPubSubServer(service, s.logger)
	interceptor := handlers.NewInterceptor(s.logger)

	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptor.RequestIdUnaryInterceptor(), interceptor.LoggingUnaryInterceptor()),
		grpc.ChainStreamInterceptor(interceptor.RequestIdStreamInterceptor(), interceptor.LoggingStreamInterceptor()),
	}
	s.server = grpc.NewServer(options...)
	proto.RegisterPubSubServer(s.server, handler)

	go func() {
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChan
		err := s.gracefulStop()

		if err != nil {
			s.logger.Fatal("Could not gracefully shutdown server", zap.Error(err))
		}
		s.logger.Info("Server has been shut down, goodbye!")
	}()
}

func (s *Server) Start() error {
	if s.server == nil {
		return errors.New("server is not initialized")
	}

	s.logger.Info("Starting server")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.config.Host, s.config.Port))
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("Listening on %s:%s", s.config.Host, s.config.Port))
	return s.server.Serve(lis)
}
