package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
	"github.com/discord-subscriptions/deploy-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/auth"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
	"github.com/discord-subscriptions/shared/health"
)

type ProvisionRequest struct {
	SubscriptionID string `json:"subscription_id"`
	UserID         string `json:"user_id"`
	GuildID        string `json:"guild_id"`
	BotType        string `json:"bot_type"`
	InstanceLabel  string `json:"instance_label,omitempty"`
	BotToken       string `json:"bot_token"`
	ImageTag       string `json:"image_tag"`
	IsZeroSetup    bool   `json:"is_zero_setup"`
}

type CustomizeRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type AddPoolTokensRequest struct {
	BotType string                `json:"bot_type"`
	Tokens  []ports.AddTokenInput `json:"tokens"`
}

type HTTPHandler struct {
	service   ports.DeploymentService
	checker   *health.Checker
	logger    *slog.Logger
	validator *auth.Validator
}

// NewHTTPHandler creates the HTTP REST controller for deployments.
func NewHTTPHandler(service ports.DeploymentService, checker *health.Checker, logger *slog.Logger, validator *auth.Validator) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		checker:   checker,
		logger:    logger,
		validator: validator,
	}
}

// RegisterRoutes registers deployment routes on the ServeMux.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	if h.checker != nil {
		mux.HandleFunc("GET /health", h.checker.HealthHandler())
		mux.HandleFunc("GET /livez", h.checker.LivezHandler())
		mux.HandleFunc("GET /readyz", h.checker.ReadyzHandler())
	} else {
		mux.HandleFunc("GET /health", h.HealthCheck)
		mux.HandleFunc("GET /livez", h.HealthCheck)
		mux.HandleFunc("GET /readyz", h.HealthCheck)
	}

	mux.HandleFunc("POST /api/v1/deployments", h.Provision)
	mux.HandleFunc("GET /api/v1/deployments/{id}", h.GetByID)
	mux.HandleFunc("GET /api/v1/deployments/guild/{guildId}", h.GetByGuild)
	mux.HandleFunc("POST /api/v1/deployments/guild/{guildId}/customize", h.Customize)
	mux.HandleFunc("POST /api/v1/deployments/{id}/customize", h.CustomizeByID)
	mux.HandleFunc("POST /api/v1/deployments/{id}/restart", h.Restart)
	mux.HandleFunc("POST /api/v1/deployments/{id}/stop", h.Stop)

	// Token pool endpoints
	mux.HandleFunc("GET /api/v1/token-pool/available", h.CheckPoolAvailable)
	mux.HandleFunc("POST /api/v1/admin/token-pool", h.AddPoolTokens)
	mux.HandleFunc("GET /api/v1/admin/token-pool", h.GetPoolStats)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "deploy-svc",
	})
}

func (h *HTTPHandler) Provision(w http.ResponseWriter, r *http.Request) {
	var req ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.GuildID == "" || req.SubscriptionID == "" || req.BotType == "" {
		h.respondError(w, http.StatusBadRequest, "Missing required deployment fields")
		return
	}

	dep, err := h.service.ProvisionBot(
		r.Context(),
		req.SubscriptionID,
		req.UserID,
		req.GuildID,
		req.BotType,
		req.InstanceLabel,
		req.BotToken,
		req.ImageTag,
		req.IsZeroSetup,
	)
	if err != nil {
		h.logger.Error("provisioning failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"data": dep,
	})
}

func (h *HTTPHandler) Customize(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildId")
	var req CustomizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" && req.AvatarURL == "" {
		h.respondError(w, http.StatusBadRequest, "Either name or avatar_url must be provided")
		return
	}

	err := h.service.CustomizeBot(r.Context(), guildID, req.Name, req.AvatarURL)
	if err != nil {
		if errors.Is(err, domain.ErrRateLimitExceeded) {
			h.respondError(w, http.StatusTooManyRequests, "Discord rate limit reached: bot name/avatar can only be changed twice per hour")
			return
		}
		h.logger.Error("bot customization failed", slog.String("guild_id", guildID), slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Bot appearance updated successfully",
	})
}

func (h *HTTPHandler) CustomizeByID(w http.ResponseWriter, r *http.Request) {
	depID := r.PathValue("id")
	var req CustomizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" && req.AvatarURL == "" {
		h.respondError(w, http.StatusBadRequest, "Either name or avatar_url must be provided")
		return
	}

	err := h.service.CustomizeBot(r.Context(), depID, req.Name, req.AvatarURL)
	if err != nil {
		if errors.Is(err, domain.ErrRateLimitExceeded) {
			h.respondError(w, http.StatusTooManyRequests, "Discord rate limit reached: bot name/avatar can only be changed twice per hour")
			return
		}
		h.logger.Error("bot customization failed", slog.String("dep_id", depID), slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Bot appearance updated successfully",
	})
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dep, err := h.service.GetDeployment(r.Context(), id)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "Deployment not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve deployment")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": dep,
	})
}

func (h *HTTPHandler) GetByGuild(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildId")
	deps, err := h.service.GetDeploymentsByGuild(r.Context(), guildID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve guild deployments")
		return
	}

	var firstDep *domain.Deployment
	for i := range deps {
		if deps[i].Status == domain.StatusRunning {
			firstDep = &deps[i]
			break
		}
	}
	if firstDep == nil && len(deps) > 0 {
		firstDep = &deps[0]
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"deployments": deps,
		"deployment":  firstDep,
		"data":        firstDep,
	})
}

func (h *HTTPHandler) Restart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.RestartBot(r.Context(), id); err != nil {
		h.logger.Error("restart failed", slog.String("id", id), slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to restart deployment")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Restart triggered successfully",
	})
}

func (h *HTTPHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.StopBot(r.Context(), id); err != nil {
		h.logger.Error("stop failed", slog.String("id", id), slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to stop deployment")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Bot deployment stopped successfully",
	})
}

func (h *HTTPHandler) CheckPoolAvailable(w http.ResponseWriter, r *http.Request) {
	botType := r.URL.Query().Get("bot_type")
	if botType == "" {
		h.respondError(w, http.StatusBadRequest, "Query parameter bot_type is required")
		return
	}

	available, count, err := h.service.CheckPoolAvailability(r.Context(), botType)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"bot_type":        botType,
		"is_available":    available,
		"available_count": count,
	})
}

func (h *HTTPHandler) AddPoolTokens(w http.ResponseWriter, r *http.Request) {
	if h.validator != nil {
		claims, err := h.validator.ExtractAndValidate(r)
		if err != nil {
			h.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		if !claims.IsAdmin && !claims.IsSuperAdmin {
			h.respondError(w, http.StatusForbidden, "Administrator privileges required")
			return
		}
	}

	var req AddPoolTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.BotType == "" || len(req.Tokens) == 0 {
		h.respondError(w, http.StatusBadRequest, "bot_type and at least one token required")
		return
	}

	added, err := h.service.AddPoolTokens(r.Context(), req.BotType, req.Tokens)
	if err != nil {
		h.logger.Error("failed adding pool tokens", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"message":     "Tokens added to pool successfully",
		"added_count": added,
	})
}

func (h *HTTPHandler) GetPoolStats(w http.ResponseWriter, r *http.Request) {
	if h.validator != nil {
		claims, err := h.validator.ExtractAndValidate(r)
		if err != nil {
			h.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		if !claims.IsAdmin && !claims.IsSuperAdmin {
			h.respondError(w, http.StatusForbidden, "Administrator privileges required")
			return
		}
	}

	stats, err := h.service.GetPoolStats(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"stats": stats,
	})
}

func (h *HTTPHandler) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *HTTPHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
