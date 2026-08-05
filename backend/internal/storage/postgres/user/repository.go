package users_repository

import (
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type Repository struct {
	pool *postgrespool.Pool
}

func NewRepository(pool *postgrespool.Pool) *Repository {
	return &Repository{pool: pool}
}
