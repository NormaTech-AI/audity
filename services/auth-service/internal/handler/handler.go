package handler

import (
	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/config"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/oidc"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/store"
	"go.uber.org/zap"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	store            *store.Store
	config           *config.Config
	jwtManager       *auth.JWTManager
	googleProvider   *oidc.OIDCProvider
	microsoftProvider *oidc.OIDCProvider
	logger           *zap.SugaredLogger
	// State storage (in production, use Redis)
	stateStore map[string]string
}

// NewHandler creates a new Handler instance
func NewHandler(
	store *store.Store,
	config *config.Config,
	jwtManager *auth.JWTManager,
	googleProvider *oidc.OIDCProvider,
	microsoftProvider *oidc.OIDCProvider,
	logger *zap.SugaredLogger,
) *Handler {
	return &Handler{
		store:            store,
		config:           config,
		jwtManager:       jwtManager,
		googleProvider:   googleProvider,
		microsoftProvider: microsoftProvider,
		logger:           logger,
		stateStore:       make(map[string]string),
	}
}
