package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/discord-subscriptions/auth-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/health"
)

type CallbackRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type DevLoginRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type HTTPHandler struct {
	service ports.AuthService
	checker *health.Checker
	logger  *slog.Logger
}

func NewHTTPHandler(service ports.AuthService, checker *health.Checker, logger *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		checker: checker,
		logger:  logger,
	}
}

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

	mux.HandleFunc("GET /api/v1/auth/discord/url", h.GetOAuthURL)
	mux.HandleFunc("POST /api/v1/auth/discord/callback", h.Callback)
	mux.HandleFunc("GET /api/v1/auth/me", h.GetProfile)
	mux.HandleFunc("GET /api/v1/auth/guilds", h.GetGuilds)
	mux.HandleFunc("POST /api/v1/auth/dev-login", h.DevLogin)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "auth-svc",
	})
}

func (h *HTTPHandler) GetOAuthURL(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = "http://localhost:3000/api/auth/callback/discord"
	}

	authURL := h.service.GetOAuthURL(redirectURI)
	h.respondJSON(w, http.StatusOK, map[string]string{
		"url": authURL,
	})
}

func (h *HTTPHandler) Callback(w http.ResponseWriter, r *http.Request) {
	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		h.respondError(w, http.StatusBadRequest, "Authorization code is required")
		return
	}

	session, err := h.service.AuthenticateWithCode(r.Context(), req.Code, req.RedirectURI)
	if err != nil {
		h.logger.Error("callback authentication failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": session,
	})
}

func (h *HTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	token := h.extractBearerToken(r)
	if token == "" {
		h.respondError(w, http.StatusUnauthorized, "Missing authorization token")
		return
	}

	user, err := h.service.GetUserSession(r.Context(), token)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": user,
	})
}

func (h *HTTPHandler) GetGuilds(w http.ResponseWriter, r *http.Request) {
	token := h.extractBearerToken(r)
	if token == "" {
		h.respondError(w, http.StatusUnauthorized, "Missing authorization token")
		return
	}

	user, err := h.service.GetUserSession(r.Context(), token)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	guilds, err := h.service.GetUserManageableGuilds(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("failed fetching manageable guilds", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve servers")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": guilds,
	})
}

func (h *HTTPHandler) DevLogin(w http.ResponseWriter, r *http.Request) {
	var req DevLoginRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	session, err := h.service.DevLogin(r.Context(), req.UserID, req.Username)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": session,
	})
}

func (h *HTTPHandler) extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
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
