package config

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabase() *pgxpool.Pool {

	connString :=
		"postgres://antiscam:antiscam@localhost:5432/antiscam?sslmode=disable"

	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Connected to PostgreSQL")

	return db
}
