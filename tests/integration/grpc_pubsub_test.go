//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"subscriber/internal/domain/dto"
	"subscriber/internal/handlers/grpc"
	grpcMocks "subscriber/internal/handlers/grpc/mocks"
	"subscriber/internal/handlers/grpc/proto"
	serviceMocks "subscriber/internal/services/mocks"
)

type mockSubscription struct {
	msgChan chan dto.Message
}

func (m *mockSubscription) Unsubscribe() {
	close(m.msgChan)
}

func TestPubSub_Subscribe(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.SubscribeRequest
		mockSetup     func(mock *serviceMocks.PubSub, stream *grpcMocks.MockServerStream)
		expectedError string
		events        []*proto.Event
	}{
		{
			name: "successful subscribe with messages",
			request: &proto.SubscribeRequest{
				Key: "test-topic",
			},
			mockSetup: func(mock *serviceMocks.PubSub, stream *grpcMocks.MockServerStream) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				stream.On("Context").Return(ctx)

				msgChan := make(chan dto.Message, 2)
				go func() {
					msgChan <- dto.Message{Topic: "test-topic", Data: "message1"}
					msgChan <- dto.Message{Topic: "test-topic", Data: "message2"}
				}()

				subscription := &mockSubscription{msgChan: msgChan}

				mock.On("Subscribe", testifyMock.Anything, testifyMock.MatchedBy(func(sub dto.Subscription) bool {
					return sub.Topic == "test-topic"
				})).Return(subscription, nil)

				stream.On("Send", testifyMock.MatchedBy(func(event *proto.Event) bool {
					return event.Data == "message1" || event.Data == "message2"
				})).Return(nil).Maybe()
			},
			events: []*proto.Event{
				{Data: "message1"},
				{Data: "message2"},
			},
			expectedError: "",
		},
		{
			name: "bus closed error",
			request: &proto.SubscribeRequest{
				Key: "test-topic",
			},
			mockSetup: func(mock *serviceMocks.PubSub, stream *grpcMocks.MockServerStream) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				stream.On("Context").Return(ctx)

				mock.On("Subscribe", testifyMock.Anything, testifyMock.Anything).Return(nil, errors.New("bus is closed"))
			},
			expectedError: "rpc error: code = Internal desc = bus is closed",
		},
		{
			name: "stream send error",
			request: &proto.SubscribeRequest{
				Key: "test-topic",
			},
			mockSetup: func(mock *serviceMocks.PubSub, stream *grpcMocks.MockServerStream) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				stream.On("Context").Return(ctx)

				msgChan := make(chan dto.Message, 1)
				go func() {
					msgChan <- dto.Message{Topic: "test-topic", Data: "message1"}
				}()

				subscription := &mockSubscription{msgChan: msgChan}

				mock.On("Subscribe", testifyMock.Anything, testifyMock.MatchedBy(func(sub dto.Subscription) bool {
					return sub.Topic == "test-topic"
				})).Return(subscription, nil)

				stream.On("Send", testifyMock.Anything).Return(errors.New("stream error")).Maybe()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPubSub := serviceMocks.NewPubSub(t)
			mockStream := grpcMocks.NewMockServerStream(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPubSub, mockStream)
			}

			logger := zap.NewNop()
			handler := grpc.NewPubSubServer(mockPubSub, logger)

			err := handler.Subscribe(tt.request, mockStream)
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPubSub_Publish(t *testing.T) {
	tests := []struct {
		name          string
		request       *proto.PublishRequest
		mockSetup     func(mock *serviceMocks.PubSub)
		expectedError string
	}{
		{
			name: "successful publish",
			request: &proto.PublishRequest{
				Key:  "test-topic",
				Data: "test-data",
			},
			mockSetup: func(mock *serviceMocks.PubSub) {
				mock.On("Publish", testifyMock.Anything, testifyMock.MatchedBy(func(msg dto.Message) bool {
					return msg.Topic == "test-topic" && msg.Data == "test-data"
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "bus closed error",
			request: &proto.PublishRequest{
				Key:  "test-topic",
				Data: "test-data",
			},
			mockSetup: func(mock *serviceMocks.PubSub) {
				mock.On("Publish", testifyMock.Anything, testifyMock.MatchedBy(func(msg dto.Message) bool {
					return msg.Topic == "test-topic" && msg.Data == "test-data"
				})).Return(errors.New("bus is closed"))
			},
			expectedError: "rpc error: code = Internal desc = bus is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPubSub := serviceMocks.NewPubSub(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockPubSub)
			}

			logger := zap.NewNop()
			handler := grpc.NewPubSubServer(mockPubSub, logger)

			_, err := handler.Publish(context.Background(), tt.request)
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
