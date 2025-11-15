package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/interview-stack/go-service/internal/config"
)

// NewPostgresPool connects to PostgreSQL using pgxpool with sane defaults.
func NewPostgresPool(ctx context.Context, pgCfg config.PostgresConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(pgCfg.DSN())
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.ConnConfig.ConnectTimeout = pgCfg.ConnectTimeout.Duration()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, pgCfg.ConnectTimeout.Duration())
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
