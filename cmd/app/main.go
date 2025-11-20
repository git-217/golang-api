package main

import (
	"context"
	"log/slog"
	"os"
	"psql_crud/internal/config"
	"psql_crud/internal/database"
	"psql_crud/internal/database/postgres"
	"psql_crud/internal/lib/logger/sl"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	logger := initLogger(cfg.Env)
	logger.Info("Initializing service", slog.String("env", cfg.Env))
	logger.Debug("Showing debug messages")

	// Inint pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("Failed to init pool", sl.Err(err))
	}
	defer dbPool.Close()

	err = postgres.InitUrlTable(ctx, dbPool)
	if err != nil {
		logger.Error("Failed to init url table", sl.Err(err))
	}
	logger.Info("Database table initialized successfully")
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
