package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NormaTech-AI/audity/services/client-service/internal/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
)

// CreateClientRequest represents the request to create a new client
type CreateClientRequest struct {
	Name        string `json:"name" validate:"required"`
	POCEmail    string `json:"poc_email" validate:"required,email"`
	EmailDomain string `json:"email_domain"`
}

// ClientResponse represents a client in API responses
type ClientResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	POCEmail  string    `json:"poc_email"`
	EmailDomain string  `json:"email_domain"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// CreateClient godoc
// @Summary Create a new client
// @Description Onboard a new client with isolated database and MinIO bucket
// @Tags clients
// @Accept json
// @Produce json
// @Param client body CreateClientRequest true "Client details"
// @Success 201 {object} ClientResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/clients [post]
func (h *Handler) CreateClient(c echo.Context) error {
	var req CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Start transaction
	var client db.Client
	err := h.store.ExecTx(ctx, func(q *db.Queries) error {
		// 1. Create client record
		var err error
		var emailDomain *string
		if req.EmailDomain != "" {
			emailDomain = &req.EmailDomain
		}
		client, err = q.CreateClient(ctx, db.CreateClientParams{
			Name:        req.Name,
			PocEmail:    req.POCEmail,
			Status:      db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumActive, Valid: true},
			EmailDomain: emailDomain,
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// 2. Provision isolated database
		dbName := fmt.Sprintf("client_%s", client.ID.String()[:8])
		dbUser := fmt.Sprintf("user_%s", client.ID.String()[:8])
		dbPassword := generateSecurePassword()

		if err := h.provisionDatabase(ctx, dbName, dbUser, dbPassword); err != nil {
			return fmt.Errorf("failed to provision database: %w", err)
		}

		// Store database credentials (password should be encrypted in production)
		_, err = q.CreateClientDatabase(ctx, db.CreateClientDatabaseParams{
			ClientID:          client.ID,
			DbName:            dbName,
			DbHost:            h.config.Database.PostgresHost,
			DbPort:            int32(h.config.Database.PostgresPort),
			DbUser:            dbUser,
			EncryptedPassword: dbPassword, // TODO: Encrypt this
		})
		if err != nil {
			return fmt.Errorf("failed to store database credentials: %w", err)
		}

		// 3. Provision MinIO bucket
		bucketName := fmt.Sprintf("client-%s", client.ID.String()[:8])
		if err := h.provisionBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("failed to provision bucket: %w", err)
		}

		// Store bucket info
		_, err = q.CreateClientBucket(ctx, db.CreateClientBucketParams{
			ClientID:   client.ID,
			BucketName: bucketName,
		})
		if err != nil {
			return fmt.Errorf("failed to store bucket info: %w", err)
		}

		return nil
	})

	if err != nil {
		h.logger.Errorw("Failed to create client", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create client",
		})
	}

	h.logger.Infow("Client created successfully", "client_id", client.ID, "name", client.Name)

	emailDomainStr := ""
	if client.EmailDomain != nil {
		emailDomainStr = *client.EmailDomain
	}

	return c.JSON(http.StatusCreated, ClientResponse{
		ID:          client.ID,
		Name:        client.Name,
		POCEmail:    client.PocEmail,
		EmailDomain: emailDomainStr,
		Status:      string(client.Status.ClientStatusEnum),
		CreatedAt:   client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// GetClient godoc
// @Summary Get client by ID
// @Description Get detailed information about a specific client
// @Tags clients
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} ClientResponse
// @Failure 404 {object} map[string]string
// @Router /api/clients/{id} [get]
func (h *Handler) GetClient(c echo.Context) error {
	idStr := c.Param("id")
	clientID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	client, err := h.store.Queries().GetClient(c.Request().Context(), clientID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Client not found",
		})
	}

	emailDomainStr := ""
	if client.EmailDomain != nil {
		emailDomainStr = *client.EmailDomain
	}

	return c.JSON(http.StatusOK, ClientResponse{
		ID:          client.ID,
		Name:        client.Name,
		POCEmail:    client.PocEmail,
		EmailDomain: emailDomainStr,
		Status:      string(client.Status.ClientStatusEnum),
		CreatedAt:   client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// ListClients godoc
// @Summary List all clients
// @Description Get a list of all clients
// @Tags clients
// @Produce json
// @Success 200 {array} ClientResponse
// @Router /api/clients [get]
func (h *Handler) ListClients(c echo.Context) error {
	clients, err := h.store.Queries().ListClients(c.Request().Context())
	if err != nil {
		h.logger.Errorw("Failed to list clients", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve clients",
		})
	}

	response := make([]ClientResponse, len(clients))
	for i, client := range clients {
		emailDomainStr := ""
		if client.EmailDomain != nil {
			emailDomainStr = *client.EmailDomain
		}
		response[i] = ClientResponse{
			ID:          client.ID,
			Name:        client.Name,
			POCEmail:    client.PocEmail,
			EmailDomain: emailDomainStr,
			Status:      string(client.Status.ClientStatusEnum),
			CreatedAt:   client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(http.StatusOK, response)
}

// provisionDatabase creates a new isolated database for a client
func (h *Handler) provisionDatabase(ctx context.Context, dbName, dbUser, dbPassword string) error {
	pool := h.store.GetPool()

	// Create database
	_, err := pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Create user
	_, err = pool.Exec(ctx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", dbUser, dbPassword))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Grant privileges
	_, err = pool.Exec(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", dbName, dbUser))
	if err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	h.logger.Infow("Database provisioned", "db_name", dbName, "db_user", dbUser)
	return nil
}

// provisionBucket creates a new MinIO bucket for a client
func (h *Handler) provisionBucket(ctx context.Context, bucketName string) error {
	exists, err := h.minio.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = h.minio.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		h.logger.Infow("Bucket created", "bucket_name", bucketName)
	}

	return nil
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	// Simple implementation - in production, use crypto/rand for better security
	return uuid.New().String()
}

// UpdateClient godoc
// @Summary Update client information
// @Description Update client details including name, POC email, email domain, and status
// @Tags clients
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param client body UpdateClientRequest true "Client update data"
// @Success 200 {object} ClientResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/clients/{id} [put]
func (h *Handler) UpdateClient(c echo.Context) error {
	idStr := c.Param("id")
	clientID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	var req UpdateClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check if client exists
	existingClient, err := h.store.Queries().GetClient(ctx, clientID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Client not found",
		})
	}

	// Prepare email domain
	var emailDomain *string
	if req.EmailDomain != "" {
		emailDomain = &req.EmailDomain
	}

	// Prepare status
	var status db.NullClientStatusEnum
	if req.Status != "" {
		switch req.Status {
		case "active":
			status = db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumActive, Valid: true}
		case "inactive":
			status = db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumInactive, Valid: true}
		case "suspended":
			status = db.NullClientStatusEnum{ClientStatusEnum: db.ClientStatusEnumSuspended, Valid: true}
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid status. Must be 'active', 'inactive', or 'suspended'",
			})
		}
	} else {
		status = existingClient.Status
	}

	// Update client
	updatedClient, err := h.store.Queries().UpdateClient(ctx, db.UpdateClientParams{
		ID:          clientID,
		Name:        req.Name,
		PocEmail:    req.POCEmail,
		Status:      status,
		EmailDomain: emailDomain,
	})
	if err != nil {
		h.logger.Errorw("Failed to update client", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update client",
		})
	}

	h.logger.Infow("Client updated successfully", "client_id", clientID, "name", updatedClient.Name)

	emailDomainStr := ""
	if updatedClient.EmailDomain != nil {
		emailDomainStr = *updatedClient.EmailDomain
	}

	return c.JSON(http.StatusOK, ClientResponse{
		ID:          updatedClient.ID,
		Name:        updatedClient.Name,
		POCEmail:    updatedClient.PocEmail,
		EmailDomain: emailDomainStr,
		Status:      string(updatedClient.Status.ClientStatusEnum),
		CreatedAt:   updatedClient.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   updatedClient.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// DeleteClient godoc
// @Summary Delete a client
// @Description Delete a client and all associated resources (database, bucket, etc.)
// @Tags clients
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/clients/{id} [delete]
func (h *Handler) DeleteClient(c echo.Context) error {
	idStr := c.Param("id")
	clientID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	ctx := c.Request().Context()

	// Check if client exists
	client, err := h.store.Queries().GetClient(ctx, clientID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Client not found",
		})
	}

	// Get client database info for cleanup
	clientDB, err := h.store.Queries().GetClientDatabase(ctx, clientID)
	var dbName, dbUser string
	if err == nil {
		dbName = clientDB.DbName
		dbUser = clientDB.DbUser
	}

	// Get client bucket info for cleanup
	clientBucket, err := h.store.Queries().GetClientBucket(ctx, clientID)
	var bucketName string
	if err == nil {
		bucketName = clientBucket.BucketName
	}

	// Start transaction to delete client and related records
	err = h.store.ExecTx(ctx, func(q *db.Queries) error {
		// Delete client frameworks
		// Note: This will cascade delete due to foreign key constraints

		// Delete client bucket record
		if err := q.DeleteClientBucket(ctx, clientID); err != nil {
			h.logger.Warnw("Failed to delete client bucket record", "error", err, "client_id", clientID)
		}

		// Delete client database record
		if err := q.DeleteClientDatabase(ctx, clientID); err != nil {
			h.logger.Warnw("Failed to delete client database record", "error", err, "client_id", clientID)
		}

		// Delete client record (this will cascade to users and other related tables)
		if err := q.DeleteClient(ctx, clientID); err != nil {
			return fmt.Errorf("failed to delete client: %w", err)
		}

		return nil
	})

	if err != nil {
		h.logger.Errorw("Failed to delete client", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete client",
		})
	}

	// Clean up infrastructure (best effort - don't fail if these fail)
	// Drop PostgreSQL database
	if dbName != "" {
		pool := h.store.GetPool()

		// Terminate connections to the database first
		_, err = pool.Exec(ctx, fmt.Sprintf(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()",
			dbName,
		))
		if err != nil {
			h.logger.Warnw("Failed to terminate database connections", "error", err, "db_name", dbName)
		}

		// Drop the database
		_, err = pool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		if err != nil {
			h.logger.Warnw("Failed to drop database", "error", err, "db_name", dbName)
		} else {
			h.logger.Infow("Database dropped", "db_name", dbName)
		}

		// Drop the database user
		if dbUser != "" {
			_, err = pool.Exec(ctx, fmt.Sprintf("DROP USER IF EXISTS %s", dbUser))
			if err != nil {
				h.logger.Warnw("Failed to drop database user", "error", err, "db_user", dbUser)
			} else {
				h.logger.Infow("Database user dropped", "db_user", dbUser)
			}
		}
	}

	// Remove MinIO bucket
	if bucketName != "" {
		// First, remove all objects in the bucket
		objectsCh := h.minio.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})

		for object := range objectsCh {
			if object.Err != nil {
				h.logger.Warnw("Error listing objects", "error", object.Err, "bucket", bucketName)
				continue
			}
			err := h.minio.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				h.logger.Warnw("Failed to remove object", "error", err, "bucket", bucketName, "object", object.Key)
			}
		}

		// Remove the bucket
		err = h.minio.RemoveBucket(ctx, bucketName)
		if err != nil {
			h.logger.Warnw("Failed to remove bucket", "error", err, "bucket_name", bucketName)
		} else {
			h.logger.Infow("Bucket removed", "bucket_name", bucketName)
		}
	}

	h.logger.Infow("Client deleted successfully", "client_id", clientID, "name", client.Name)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Client deleted successfully",
	})
}

// UpdateClientRequest represents the request to update a client
type UpdateClientRequest struct {
	Name        string `json:"name" validate:"required"`
	POCEmail    string `json:"poc_email" validate:"required,email"`
	EmailDomain string `json:"email_domain"`
	Status      string `json:"status"`
}
