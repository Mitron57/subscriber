package integration

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"sync"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"subscriber/config"
	"subscriber/internal/handlers/grpc/proto"
	"subscriber/internal/server"
)

const (
	Host = "0.0.0.0"
	Port = "50051"
)

type testClient struct {
	inner proto.PubSubClient
	conn  *grpc.ClientConn
}

var (
	srv    *server.Server
	client testClient
)

func Setup() {
	logger := zap.NewNop()
	cfg := config.Config{Host: Host, Port: Port}

	srv = server.NewServer(&cfg, logger)
	srv.Init()
	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()

	host := fmt.Sprintf("%s:%s", Host, Port)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	client.conn, _ = grpc.NewClient(host, opts...)
	client.inner = proto.NewPubSubClient(client.conn)
}

func Teardown() {
	err := srv.GracefulStop()
	if err != nil {
		panic(err)
	}

	err = client.conn.Close()
	if err != nil {
		panic(err)
	}
}

func TestSubscriber(t *testing.T) {
	const topic = "test"
	ctx, cancel := context.WithCancel(context.Background())

	param := proto.SubscribeRequest{Key: topic}
	stream, err := client.inner.Subscribe(ctx, &param)
	if err != nil {
		t.Error(err)
	}

	messages := map[string]struct{}{
		"1": {},
		"2": {},
		"3": {},
	}

	errs := make(chan error, len(messages))
	var wg sync.WaitGroup
	wg.Add(len(messages))
	for message := range messages {
		go func() {
			defer wg.Done()
			_, err := client.inner.Publish(ctx, &proto.PublishRequest{Key: topic, Data: message})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for i := 0; i < len(messages); i++ {
		err, ok := <-errs
		if !ok {
			break
		}
		if err != nil {
			t.Errorf("publish error: %v", err)
		}
	}

	received := make(map[string]struct{}, len(messages))

	for err != io.EOF && !maps.Equal(received, messages) {
		var msg *proto.Event
		msg, err = stream.Recv()
		if err != nil {
			t.Error(err)
			continue
		}
		received[msg.Data] = struct{}{}
	}
	cancel()
}

func TestMain(m *testing.M) {
	Setup()
	code := m.Run()
	Teardown()
	os.Exit(code)
}
