package repo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *URLRepo {
	return &URLRepo{pool: pool}
}
