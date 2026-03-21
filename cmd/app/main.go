package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"psql_crud/internal/config"
	"psql_crud/internal/http-server/handlers/redirect"
	del "psql_crud/internal/http-server/handlers/url/delete"
	"psql_crud/internal/http-server/handlers/url/save"
	limiterMware "psql_crud/internal/http-server/middleware/limiter"
	mwLog "psql_crud/internal/http-server/middleware/logger"
	"psql_crud/internal/lib/limiter"
	"psql_crud/internal/lib/logger/handlers/slogpretty"
	"psql_crud/internal/lib/logger/sl"
	"psql_crud/internal/storage/postgres"
	db_req "psql_crud/internal/storage/postgres"
	pool "psql_crud/internal/storage/postgres/pgx"
	"syscall"
	"time"

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
	logger.Info("initializing service", slog.String("env", cfg.Env))
	logger.Debug("showing debug messages")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pool.New(ctx, cfg)
	if err != nil {
		logger.Error("Failed to init pool", sl.Err(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	if err = postgres.New(ctx, dbPool); err != nil {
		logger.Error("Failed to init url table", sl.Err(err))
	}

	repo := db_req.NewURLRepo(dbPool)

	ipLimiter := limiter.NewIPRateLimiter(cfg)

	router := chi.NewRouter()
	router.Use(limiterMware.RateLimiterMiddleware(ipLimiter))
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLog.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url/new", save.New(logger, repo))
	router.Get("/{alias}", redirect.New(logger, repo))
	router.Delete("/{alias}", del.One(logger, repo))

	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.IdleTimeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}
	go func() {
		logger.Info("starting server", slog.String("address", cfg.HttpServer.Address))
		if err := server.ListenAndServe(); err != nil {
			logger.Error("failed to run server", sl.Err(err))
		}
		logger.Error("server listen error")
	}()

	<-ctx.Done()
	logger.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", sl.Err(err))
	} else {
		logger.Info("server stopped correctly")
	}
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
