package database

import (
	"context"
	"fmt"
	"psql_crud/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	const op = "internal.repository.repository.NewPool"

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.PsqlConn.User,
		cfg.PsqlConn.Password,
		cfg.PsqlConn.Host,
		cfg.PsqlConn.Port,
		cfg.PsqlConn.DbName,
		cfg.PsqlConn.SSLMode,
	)
	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	poolCfg.MaxConns = cfg.PsqlConn.MaxConns
	poolCfg.MinConns = cfg.PsqlConn.MinConns
	poolCfg.MaxConnLifetime = time.Duration(cfg.PsqlConn.ConnLife) * time.Hour
	poolCfg.MaxConnIdleTime = time.Duration(cfg.PsqlConn.ConnIdle) * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return pool, nil
}