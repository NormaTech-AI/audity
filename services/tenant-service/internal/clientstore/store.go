package clientstore

import (
	"context"
	"fmt"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/crypto"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ClientStore manages connections to client-specific databases
type ClientStore struct {
	tenantStore *db.Queries
	encryptor   *crypto.Encryptor
	logger      *zap.SugaredLogger
	// Cache for client database connections
	connectionCache map[uuid.UUID]*pgxpool.Pool
}

// NewClientStore creates a new client store manager
func NewClientStore(tenantStore *db.Queries, encryptor *crypto.Encryptor, logger *zap.SugaredLogger) *ClientStore {
	return &ClientStore{
		tenantStore:     tenantStore,
		encryptor:       encryptor,
		logger:          logger,
		connectionCache: make(map[uuid.UUID]*pgxpool.Pool),
	}
}

// GetClientQueries returns a Queries instance for a specific client's database
func (cs *ClientStore) GetClientQueries(ctx context.Context, clientID uuid.UUID) (*clientdb.Queries, *pgxpool.Pool, error) {
	// Check cache first
	if pool, exists := cs.connectionCache[clientID]; exists {
		return clientdb.New(pool), pool, nil
	}

	// Get client database credentials from tenant_db
	clientDB, err := cs.tenantStore.GetClientDatabase(ctx, clientID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client database info: %w", err)
	}

	// Decrypt password
	password, err := cs.encryptor.Decrypt(clientDB.EncryptedPassword)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Build connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		clientDB.DbUser, password, clientDB.DbHost, clientDB.DbPort, clientDB.DbName)

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to ping client database: %w", err)
	}

	// Cache the connection
	cs.connectionCache[clientID] = pool

	cs.logger.Infow("Connected to client database", "client_id", clientID, "db_name", clientDB.DbName)

	return clientdb.New(pool), pool, nil
}

// CloseClientConnection closes and removes a cached connection
func (cs *ClientStore) CloseClientConnection(clientID uuid.UUID) {
	if pool, exists := cs.connectionCache[clientID]; exists {
		pool.Close()
		delete(cs.connectionCache, clientID)
		cs.logger.Infow("Closed client database connection", "client_id", clientID)
	}
}

// CloseAll closes all cached connections
func (cs *ClientStore) CloseAll() {
	for clientID, pool := range cs.connectionCache {
		pool.Close()
		cs.logger.Infow("Closed client database connection", "client_id", clientID)
	}
	cs.connectionCache = make(map[uuid.UUID]*pgxpool.Pool)
}

// ExecClientTx executes a function within a database transaction on a client database
func (cs *ClientStore) ExecClientTx(ctx context.Context, clientID uuid.UUID, fn func(*clientdb.Queries) error) error {
	_, pool, err := cs.GetClientQueries(ctx, clientID)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := clientdb.New(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
