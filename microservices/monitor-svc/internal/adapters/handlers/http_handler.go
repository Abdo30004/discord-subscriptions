package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/monitor-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/health"
)

type RegisterTargetRequest struct {
	BotID           string `json:"bot_id"`
	GuildID         string `json:"guild_id"`
	InstanceLabel   string `json:"instance_label,omitempty"`
	HealthURL       string `json:"health_url"`
	PollIntervalSec int    `json:"poll_interval_sec"`
}

type HTTPHandler struct {
	service ports.MonitorService
	checker *health.Checker
	logger  *slog.Logger
}

// NewHTTPHandler creates the HTTP REST controller for monitoring endpoints.
func NewHTTPHandler(service ports.MonitorService, checker *health.Checker, logger *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		checker: checker,
		logger:  logger,
	}
}

// RegisterRoutes registers the HTTP endpoints on the ServeMux.
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

	mux.HandleFunc("GET /api/v1/monitor/targets", h.ListTargets)
	mux.HandleFunc("GET /api/v1/monitor/guild/{guildId}", h.GetGuildStatus)
	mux.HandleFunc("POST /api/v1/monitor/targets", h.RegisterTarget)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "monitor-svc",
	})
}

func (h *HTTPHandler) ListTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := h.service.ListTargets(r.Context())
	if err != nil {
		h.logger.Error("failed listing targets", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve monitoring targets")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": targets,
	})
}

func (h *HTTPHandler) GetGuildStatus(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildId")
	targets, err := h.service.GetGuildTargets(r.Context(), guildID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve monitoring targets")
		return
	}

	var firstTarget *domain.MonitoringTarget
	var logs []domain.CheckResult
	if len(targets) > 0 {
		firstTarget = &targets[0]
		_, logs, _ = h.service.GetGuildStatus(r.Context(), guildID)
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"targets":     targets,
			"target":      firstTarget,
			"recent_logs": logs,
		},
	})
}

func (h *HTTPHandler) RegisterTarget(w http.ResponseWriter, r *http.Request) {
	var req RegisterTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	target, err := h.service.RegisterTarget(r.Context(), req.BotID, req.GuildID, req.InstanceLabel, req.HealthURL, req.PollIntervalSec)
	if err != nil {
		h.logger.Error("failed registering target", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"data": target,
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
