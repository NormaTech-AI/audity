package handler

import (
	"github.com/NormaTech-AI/audity/services/framework-service/internal/config"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/store"
	"go.uber.org/zap"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	store  *store.Store
	config *config.Config
	logger *zap.SugaredLogger
}

// NewHandler creates a new Handler instance
func NewHandler(
	store *store.Store,
	config *config.Config,
	logger *zap.SugaredLogger,
) *Handler {
	return &Handler{
		store:  store,
		config: config,
		logger: logger,
	}
}
