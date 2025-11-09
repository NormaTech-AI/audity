package handler

import (
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientstore"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/crypto"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/framework"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/migrations"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/store"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	store                  *store.Store
	config                 *config.Config
	encryptor              *crypto.Encryptor
	minio                  *minio.Client
	logger                 *zap.SugaredLogger
	clientMigrationRunner  *migrations.ClientMigrationRunner
	clientStore            *clientstore.ClientStore
	frameworkService       *framework.Service
}

// NewHandler creates a new Handler instance
func NewHandler(
	store *store.Store,
	config *config.Config,
	encryptor *crypto.Encryptor,
	minioClient *minio.Client,
	logger *zap.SugaredLogger,
	clientMigrationRunner *migrations.ClientMigrationRunner,
	clientStore *clientstore.ClientStore,
	frameworkService *framework.Service,
) *Handler {
	return &Handler{
		store:                 store,
		config:                config,
		encryptor:             encryptor,
		minio:                 minioClient,
		logger:                logger,
		clientMigrationRunner: clientMigrationRunner,
		clientStore:           clientStore,
		frameworkService:      frameworkService,
	}
}
