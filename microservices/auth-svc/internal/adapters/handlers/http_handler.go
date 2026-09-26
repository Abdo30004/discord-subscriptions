package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/discord-subscriptions/auth-svc/internal/core/ports"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
	"github.com/discord-subscriptions/shared/health"
)

type CallbackRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type PromoteAdminRequest struct {
	DiscordID string `json:"discord_id"`
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
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.GetProfile)
	mux.HandleFunc("GET /api/v1/auth/guilds", h.GetGuilds)
	mux.HandleFunc("GET /api/v1/auth/admins", h.ListAdmins)
	mux.HandleFunc("GET /api/v1/auth/users/search", h.SearchUsers)
	mux.HandleFunc("POST /api/v1/auth/admins", h.PromoteAdmin)
	mux.HandleFunc("DELETE /api/v1/auth/admins/{discordId}", h.RevokeAdmin)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": session,
	})
}

func (h *HTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
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

func (h *HTTPHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
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

	admins, err := h.service.ListAdmins(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrUnauthorized) {
			h.respondError(w, http.StatusForbidden, "Forbidden: administrator privileges required")
			return
		}
		h.logger.Error("failed listing admins", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to list admins")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": admins,
	})
}

func (h *HTTPHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query().Get("q")
	users, err := h.service.SearchUsers(r.Context(), user.ID, query)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrUnauthorized) {
			h.respondError(w, http.StatusForbidden, "Forbidden: super administrator privileges required")
			return
		}
		h.logger.Error("failed searching users", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": users,
	})
}

func (h *HTTPHandler) PromoteAdmin(w http.ResponseWriter, r *http.Request) {
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

	var req PromoteAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DiscordID) == "" {
		h.respondError(w, http.StatusBadRequest, "Invalid request body: discord_id is required")
		return
	}

	if err := h.service.PromoteAdmin(r.Context(), user.ID, strings.TrimSpace(req.DiscordID)); err != nil {
		if errors.Is(err, sharedErrors.ErrUnauthorized) {
			h.respondError(w, http.StatusForbidden, "Forbidden: only Super Admins can promote administrators")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message": "User successfully promoted to administrator",
	})
}

func (h *HTTPHandler) RevokeAdmin(w http.ResponseWriter, r *http.Request) {
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

	targetDiscordID := strings.TrimSpace(r.PathValue("discordId"))
	if targetDiscordID == "" {
		h.respondError(w, http.StatusBadRequest, "Target discord ID is required in URL path")
		return
	}

	if err := h.service.RevokeAdmin(r.Context(), user.ID, targetDiscordID); err != nil {
		if errors.Is(err, sharedErrors.ErrUnauthorized) {
			h.respondError(w, http.StatusForbidden, "Forbidden: only Super Admins can revoke administrators")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message": "Admin privileges revoked successfully",
	})
}

func (h *HTTPHandler) extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	if cookie, err := r.Cookie("auth_token"); err == nil && cookie.Value != "" {
		return cookie.Value
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
