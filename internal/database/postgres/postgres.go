package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitUrlTable(ctx context.Context, pool *pgxpool.Pool) error {
	const op = "interanl.database.postgres.InitUrlTable"
	query := `
		CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		alias TEXT UNIQUE NOT NULL,
		url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_url_alias ON url(alias);
	`
	_, err = pool.Exec(ctx, indexQuery)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}
