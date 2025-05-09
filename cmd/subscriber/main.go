package main

import (
	"flag"

	"go.uber.org/zap"

	"subscriber/config"
	"subscriber/internal/server"
)

func main() {
	configPath := flag.String("c", "./config/config.yaml", "config path")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)

	logger := zap.Must(zap.NewProduction())
	if err != nil {
		logger.Fatal(err.Error())
	}

	grpcServer := server.NewServer(cfg, logger)
	grpcServer.Init()
	if err = grpcServer.Start(); err != nil {
		logger.Fatal(err.Error())
	}
}
