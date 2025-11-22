package main

import (
	"context"
	"log/slog"
	"os"
	"psql_crud/internal/config"
	"psql_crud/internal/lib/logger/sl"
	"psql_crud/internal/storage/postgres"
	pool "psql_crud/internal/storage/postgres/pgx"
	"time"
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

	dbPool, err := pool.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("Failed to init pool", sl.Err(err))
	}
	defer dbPool.Close()

	err = postgres.InitUrlTable(ctx, dbPool)
	if err != nil {
		logger.Error("Failed to init url table", sl.Err(err))
	}
	logger.Info("Database table initialized successfully")

	time.Sleep(time.Second)
	logger.Info("Adding a row into db: 'ex.com', 'check'")
	r := postgres.NewURLRepo(dbPool)
	res, err := r.SaveURL(ctx, "ex.com", "check123")
	if err != nil {
		logger.Error("Failed to insert data", sl.Err(err))
	}
	_, err = r.GetURL(ctx, "asdggasd")
	if err != nil {
		logger.Error("Failed to get url.", sl.Err(err))
	}
	logger.Info("Added a row", "id", slog.IntValue(res))

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
