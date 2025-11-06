package pgdb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConns          = 12
	minConns          = 0
	connMaxLifetime   = 5 * time.Minute
	maxConnIdleTime   = 30 * time.Minute
	healthCheckPeriod = 1 * time.Minute
	connectTimeout    = 30 * time.Second
	sslMode           = "disable"
)

type Config interface {
	Host() string
	Port() string
	User() string
	Password() string
	DB() string
}

func New(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	dsn := newDSN(config)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = connMaxLifetime
	poolConfig.MaxConnIdleTime = maxConnIdleTime
	poolConfig.HealthCheckPeriod = healthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = connectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func newDSN(config Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User(),
		config.Password(),
		config.Host(),
		config.Port(),
		config.DB(),
		sslMode,
	)
}
