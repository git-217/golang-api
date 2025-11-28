package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"psql_crud/internal/config"
	"psql_crud/internal/http-server/handlers/url/save"
	mwLog "psql_crud/internal/http-server/middleware/logger"
	"psql_crud/internal/lib/logger/handlers/slogpretty"
	"psql_crud/internal/lib/logger/sl"
	"psql_crud/internal/storage/postgres"
	db_req "psql_crud/internal/storage/postgres"
	pool "psql_crud/internal/storage/postgres/pgx"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := pool.New(ctx, cfg)
	if err != nil {
		logger.Error("Failed to init pool", sl.Err(err))
	}
	defer dbPool.Close()

	if err = postgres.New(ctx, dbPool); err != nil {
		logger.Error("Failed to init url table", sl.Err(err))
	}

	repo := db_req.NewURLRepo(dbPool)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLog.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url/new", save.New(logger, repo))

	logger.Info("starging server", slog.String("address", cfg.Http_server.Address))

	server := &http.Server{
		Addr:         cfg.Http_server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Http_server.IdleTimeout,
		WriteTimeout: cfg.Http_server.Timeout,
		IdleTimeout:  cfg.Http_server.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("failed to run server", sl.Err(err))
	}
	logger.Error("server stopped")
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	hanlder := opts.NewPrettyHandler(os.Stdout)
	return slog.New(hanlder)
}
