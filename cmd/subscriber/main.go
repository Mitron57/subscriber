package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"subscriber/config"
	"subscriber/internal/server"
)

func main() {
	logger := zap.Must(zap.NewProduction())
	if err := godotenv.Load(); err != nil {
		logger.Warn("Error loading .env file")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}

	grpcServer := server.NewServer(cfg, logger)
	grpcServer.Init()
	if err = grpcServer.Start(); err != nil {
		logger.Fatal(err.Error())
	}
}
