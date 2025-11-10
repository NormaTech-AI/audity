package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ClientDBInfo holds database connection information for a client
type ClientDBInfo struct {
	ClientID          uuid.UUID
	DBName            string
	DBHost            string
	DBPort            int32
	DBUser            string
	EncryptedPassword string
}

// Decryptor interface for password decryption
type Decryptor interface {
	Decrypt(encrypted string) (string, error)
}

// ClientPoolCache manages connection pools for client databases
type ClientPoolCache struct {
	tenantPool *pgxpool.Pool
	decryptor  Decryptor
	logger     *zap.SugaredLogger
	mu         sync.RWMutex
	pools      map[uuid.UUID]*pgxpool.Pool
}

// NewClientPoolCache creates a new client pool cache
func NewClientPoolCache(tenantPool *pgxpool.Pool, decryptor Decryptor, logger *zap.SugaredLogger) *ClientPoolCache {
	return &ClientPoolCache{
		tenantPool: tenantPool,
		decryptor:  decryptor,
		logger:     logger,
		pools:      make(map[uuid.UUID]*pgxpool.Pool),
	}
}

// GetClientPool returns a connection pool for the specified client
// It caches the connection and reuses it for subsequent requests
func (c *ClientPoolCache) GetClientPool(ctx context.Context, clientID uuid.UUID) (*pgxpool.Pool, error) {
	// Check cache first with read lock
	c.mu.RLock()
	if pool, exists := c.pools[clientID]; exists {
		c.mu.RUnlock()
		// Verify pool is still healthy
		if pingErr := pool.Ping(ctx); pingErr == nil {
			return pool, nil
		} else {
			// Pool is unhealthy, remove it and create new one
			c.logger.Warnw("Client pool unhealthy, recreating", "client_id", clientID, "error", pingErr)
			c.removePool(clientID)
		}
	} else {
		c.mu.RUnlock()
	}

	// Get write lock to create new pool
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check if another goroutine created the pool
	if pool, exists := c.pools[clientID]; exists {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		}
		// Pool is unhealthy, close and remove it
		pool.Close()
		delete(c.pools, clientID)
	}

	// Fetch client database info from tenant_db
	dbInfo, err := c.getClientDBInfo(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client database info: %w", err)
	}

	// Decrypt password
	password, err := c.decryptor.Decrypt(dbInfo.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Build connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbInfo.DBUser, password, dbInfo.DBHost, dbInfo.DBPort, dbInfo.DBName)

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping client database: %w", err)
	}

	// Cache the connection
	c.pools[clientID] = pool

	c.logger.Infow("Created client database connection pool",
		"client_id", clientID,
		"db_name", dbInfo.DBName,
		"db_host", dbInfo.DBHost)

	return pool, nil
}

// getClientDBInfo fetches client database credentials from tenant_db
func (c *ClientPoolCache) getClientDBInfo(ctx context.Context, clientID uuid.UUID) (*ClientDBInfo, error) {
	query := `
		SELECT client_id, db_name, db_host, db_port, db_user, encrypted_password
		FROM client_databases
		WHERE client_id = $1
	`

	var info ClientDBInfo
	err := c.tenantPool.QueryRow(ctx, query, clientID).Scan(
		&info.ClientID,
		&info.DBName,
		&info.DBHost,
		&info.DBPort,
		&info.DBUser,
		&info.EncryptedPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query client database info: %w", err)
	}

	return &info, nil
}

// removePool removes a pool from cache (must be called with write lock or use RemovePool)
func (c *ClientPoolCache) removePool(clientID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if pool, exists := c.pools[clientID]; exists {
		pool.Close()
		delete(c.pools, clientID)
		c.logger.Infow("Removed client database connection pool", "client_id", clientID)
	}
}

// RemovePool removes and closes a specific client's connection pool
func (c *ClientPoolCache) RemovePool(clientID uuid.UUID) {
	c.removePool(clientID)
}

// Close closes all cached connection pools
func (c *ClientPoolCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for clientID, pool := range c.pools {
		pool.Close()
		c.logger.Infow("Closed client database connection pool", "client_id", clientID)
	}
	c.pools = make(map[uuid.UUID]*pgxpool.Pool)
}

// GetPoolCount returns the number of cached pools (for monitoring)
func (c *ClientPoolCache) GetPoolCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.pools)
}
