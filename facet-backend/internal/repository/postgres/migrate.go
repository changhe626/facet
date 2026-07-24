package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/000001_init_schema.up.sql
var initSchema string

// RunMigrations executes the database migration scripts on startup.
// Uses idempotent SQL (IF NOT EXISTS) so it's safe to run multiple times.
func RunMigrations(pool *pgxpool.Pool) error {
	slog.Info("running database migrations...")

	// First ensure the facet schema exists
	_, err := pool.Exec(context.Background(), "CREATE SCHEMA IF NOT EXISTS facet")
	if err != nil {
		return fmt.Errorf("failed to create facet schema: %w", err)
	}

	// Run the complete init schema
	_, err = pool.Exec(context.Background(), initSchema)
	if err != nil {
		return fmt.Errorf("failed to run init schema migration: %w", err)
	}

	slog.Info("database migrations completed")
	return nil
}