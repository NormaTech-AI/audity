package store

import (
	"context"

	"github.com/NormaTech-AI/audity/services/client-service/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps the database connection pool and sqlc queries
type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewStore creates a new Store instance
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: db.New(pool),
	}
}

// GetPool returns the underlying connection pool
func (s *Store) GetPool() *pgxpool.Pool {
	return s.pool
}

// Queries returns the sqlc queries interface
func (s *Store) Queries() *db.Queries {
	return s.queries
}

// Close closes the database connection pool
func (s *Store) Close() {
	s.pool.Close()
}

// Ping checks if the database is reachable
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Transaction helper methods
func (s *Store) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	q := db.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
