package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger interface {
	Error(ctx context.Context, msg string)
}

type Storage struct {
	logger Logger
	db     *pgxpool.Pool
}

func New(logger Logger, db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}
