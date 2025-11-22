package postgres

import (
	"context"
	"fmt"
	"psql_crud/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomURL struct {
	Id    string `json:"id"`
	URL   string `json:"original_url"`
	Alias string `json:"alias"`
}

type URLRepo struct {
	pool *pgxpool.Pool
}

func InitUrlTable(ctx context.Context, pool *pgxpool.Pool) error {
	const op = "interanl.database.postgres.InitUrlTable"
	query := `
		CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		alias TEXT UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_url_alias ON urls(alias);
	`
	_, err = pool.Exec(ctx, indexQuery)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func NewURLRepo(pool *pgxpool.Pool) *URLRepo {
	return &URLRepo{pool: pool}
}

func (r *URLRepo) SaveURL(ctx context.Context, urlToSave string, alias string) (int, error) {
	const op = "internal.storage.postgres.SaveURL"

	var id int
	err := r.pool.QueryRow(ctx, `INSERT INTO urls(original_url, alias) 
								values ($1, $2)
								RETURNING id`,
		urlToSave, alias).Scan(&id)

	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return 0, fmt.Errorf("%s: %v", op, storage.ErrURLExists)
			}
		}
		return 0, fmt.Errorf("%s: failed to save url. %s: %v", op, urlToSave, err)
	}

	return id, nil
}

func (r *URLRepo) GetURL(ctx context.Context, alias string) (*CustomURL, error) {
	const op = "internal.storage.postgres.GetURL"

	var URLData CustomURL
	err := r.pool.QueryRow(ctx,
		`SELECT id, original_url, alias FROM urls where alias=$1`,
		alias).Scan(&URLData.Id, &URLData.URL, &URLData.Alias)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %v", op, storage.ErrURLNotFound)
		}
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return &URLData, nil
}

func (r *URLRepo) DeleteURL(ctx context.Context, alias string) error {
	const op = "internal.storage.postgres.DeleteURL"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM urls WHERE alias=$1`, alias)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return tx.Commit(ctx)
}

func (r *URLRepo) UpdateURLAlias(ctx context.Context, old_alias string, new_alias string) error {
	const op = "internal.storage.postgres.UpdateURLAlias"

	update_time := time.Now().Format("2006-01-02 15:04:05")

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE urls SET (alias=$1, updated_at=$2) WHERE alias=$3`,
		new_alias,
		update_time,
		old_alias,
	)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return tx.Commit(ctx)
}
