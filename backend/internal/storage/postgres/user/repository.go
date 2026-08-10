package users_repository

import (
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type UsersRepository struct {
	pool *postgrespool.Pool
}

func NewRepository(pool *postgrespool.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}
