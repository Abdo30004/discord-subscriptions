package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/discord-subscriptions/catalog-svc/internal/core/ports"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
	"github.com/discord-subscriptions/shared/health"
)

type HTTPHandler struct {
	catalogSvc ports.CatalogService
	checker    *health.Checker
	logger     *slog.Logger
}

// NewHTTPHandler constructs the REST controller for catalog endpoints.
func NewHTTPHandler(catalogSvc ports.CatalogService, checker *health.Checker, logger *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		catalogSvc: catalogSvc,
		checker:    checker,
		logger:     logger,
	}
}

// RegisterRoutes registers the catalog endpoints onto the provided ServeMux.
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

	mux.HandleFunc("GET /api/v1/catalog/bots", h.ListBots)
	mux.HandleFunc("GET /api/v1/catalog/bots/{idOrSlug}", h.GetBot)
	mux.HandleFunc("GET /api/v1/catalog/plans/{id}", h.GetPlan)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "catalog-svc",
	})
}

func (h *HTTPHandler) ListBots(w http.ResponseWriter, r *http.Request) {
	bots, err := h.catalogSvc.ListBots(r.Context())
	if err != nil {
		h.logger.Error("failed listing bots", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve bot catalog")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": bots,
	})
}

func (h *HTTPHandler) GetBot(w http.ResponseWriter, r *http.Request) {
	idOrSlug := r.PathValue("idOrSlug")
	if idOrSlug == "" {
		h.respondError(w, http.StatusBadRequest, "Missing bot identifier")
		return
	}

	bot, err := h.catalogSvc.GetBotDetails(r.Context(), idOrSlug)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "Bot not found")
			return
		}
		h.logger.Error("failed getting bot", slog.String("identifier", idOrSlug), slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch bot details")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": bot,
	})
}

func (h *HTTPHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "Missing plan identifier")
		return
	}

	plan, err := h.catalogSvc.GetPlan(r.Context(), id)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "Plan not found")
			return
		}
		h.logger.Error("failed getting plan", slog.String("id", id), slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch plan details")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": plan,
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
