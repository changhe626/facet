package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresDB creates and returns a pgx native connection pool
func NewPostgresDB(dsn string) (*pgxpool.Pool, error) {
	// 1. Create a native connection pool using pgxpool.New
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	// 2. Verify the connection is successful
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	slog.Info("successfully connected to postgres database using pgxpool")
	return pool, nil
}
