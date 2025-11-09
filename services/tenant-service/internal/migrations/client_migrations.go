package migrations

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// ClientMigrationRunner handles migrations for client-specific databases
type ClientMigrationRunner struct {
	migrationsPath string
	logger         *zap.SugaredLogger
}

// NewClientMigrationRunner creates a new migration runner
func NewClientMigrationRunner(migrationsPath string, logger *zap.SugaredLogger) *ClientMigrationRunner {
	return &ClientMigrationRunner{
		migrationsPath: migrationsPath,
		logger:         logger,
	}
}

// RunMigrations executes all pending migrations on a client database
func (r *ClientMigrationRunner) RunMigrations(ctx context.Context, dbURL string) error {
	r.logger.Infow("Running client database migrations", "db_url", maskPassword(dbURL))

	m, err := migrate.New(
		fmt.Sprintf("file://%s", r.migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	// Run all pending migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			r.logger.Info("No migrations to run - database is up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		r.logger.Warnw("Could not get migration version", "error", err)
	} else {
		r.logger.Infow("Migrations completed successfully", "version", version, "dirty", dirty)
	}

	return nil
}

// RollbackMigrations rolls back the last migration
func (r *ClientMigrationRunner) RollbackMigrations(ctx context.Context, dbURL string) error {
	r.logger.Infow("Rolling back client database migrations", "db_url", maskPassword(dbURL))

	m, err := migrate.New(
		fmt.Sprintf("file://%s", r.migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	// Roll back one step
	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			r.logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		r.logger.Warnw("Could not get migration version", "error", err)
	} else {
		r.logger.Infow("Rollback completed successfully", "version", version, "dirty", dirty)
	}

	return nil
}

// maskPassword masks the password in a database URL for logging
func maskPassword(dbURL string) string {
	// Simple masking for security - hide password in logs
	return "postgres://***:***@***"
}
