package postgres

import (
	"psql_crud/internal/storage"
	"time"
	"fmt"
	"context"
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

func NewURLRepo(pool *pgxpool.Pool) *URLRepo {
	return &URLRepo{pool: pool}
}

func (r *URLRepo) SaveURL(ctx context.Context, urlToSave string, alias string) (int, error) {
	const op = "internal.storage.postgres.SaveURL"

	time.Sleep(5*time.Second)
	var id int
	err := r.pool.QueryRow(ctx, `INSERT INTO urls(original_url, alias) 
								values ($1, $2)
								RETURNING id`,
		urlToSave, alias).Scan(&id)

	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
			}
		}
		return 0, fmt.Errorf("%s: failed to save url. %s: %w", op, urlToSave, err)
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
			return nil, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &URLData, nil
}

func (r *URLRepo) DeleteURL(ctx context.Context, alias string) error {
	const op = "internal.storage.postgres.DeleteURL"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM urls WHERE alias=$1`, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit(ctx)
}

func (r *URLRepo) UpdateURLAlias(ctx context.Context, old_alias string, new_alias string) error {
	const op = "internal.storage.postgres.UpdateURLAlias"

	update_time := time.Now().Format("2006-01-02 15:04:05")

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE urls SET (alias=$1, updated_at=$2) WHERE alias=$3`,
		new_alias,
		update_time,
		old_alias,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit(ctx)
}
