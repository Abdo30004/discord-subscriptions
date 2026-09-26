package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	"github.com/discord-subscriptions/billing-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/auth"
	"github.com/discord-subscriptions/shared/health"
	"github.com/google/uuid"
)

type RedeemRequest struct {
	Code    string `json:"code"`
	UserID  string `json:"user_id"`
	GuildID string `json:"guild_id"`
}

type AdminGrantRequest struct {
	UserID        string `json:"user_id"`
	GuildID       string `json:"guild_id"`
	BotType       string `json:"bot_type"`
	PlanID        string `json:"plan_id"`
	InstanceLabel string `json:"instance_label,omitempty"`
	DurationDays  int    `json:"duration_days"`
	IsDedicated   bool   `json:"is_dedicated"`
	IsZeroSetup   bool   `json:"is_zero_setup"`
}

type CreatePromoRequest struct {
	Code          string `json:"code"`
	DiscountType  string `json:"discount_type"` // "percentage" or "fixed"
	DiscountValue int64  `json:"discount_value"`
	MaxUses       int    `json:"max_uses"`
	DurationDays  int    `json:"duration_days,omitempty"`
}

type CreateVoucherRequest struct {
	Code         string `json:"code,omitempty"`
	PlanID       string `json:"plan_id"`
	BotType      string `json:"bot_type"`
	DurationDays int    `json:"duration_days"`
	IsDedicated  bool   `json:"is_dedicated"`
}

type HTTPHandler struct {
	service   ports.BillingService
	checker   *health.Checker
	logger    *slog.Logger
	validator *auth.Validator
}

func NewHTTPHandler(service ports.BillingService, checker *health.Checker, logger *slog.Logger, validator *auth.Validator) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		checker:   checker,
		logger:    logger,
		validator: validator,
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

	mux.HandleFunc("POST /api/v1/billing/checkout", h.Checkout)
	mux.HandleFunc("POST /api/v1/billing/redeem", h.Redeem)
	mux.HandleFunc("POST /api/v1/billing/admin/grant", h.AdminGrant)
	mux.HandleFunc("POST /api/v1/billing/admin/promo", h.CreatePromo)
	mux.HandleFunc("POST /api/v1/billing/admin/voucher", h.CreateVoucher)
	mux.HandleFunc("GET /api/v1/billing/subscriptions/guild/{guildId}", h.GetGuildSubscription)
	mux.HandleFunc("POST /api/v1/billing/webhooks/paypal", h.HandlePayPalWebhook)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "billing-svc",
	})
}

func (h *HTTPHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req ports.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.InitiateCheckout(r.Context(), req)
	if err != nil {
		h.logger.Error("checkout initiation failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"data": resp,
	})
}

func (h *HTTPHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sub, err := h.service.RedeemVoucher(r.Context(), req.Code, req.UserID, req.GuildID)
	if err != nil {
		h.logger.Error("voucher redemption failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":      "Voucher redeemed successfully",
		"subscription": sub,
	})
}

func (h *HTTPHandler) AdminGrant(w http.ResponseWriter, r *http.Request) {
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

	var req AdminGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sub, err := h.service.AdminGrantSubscription(
		r.Context(),
		req.UserID, req.GuildID, req.BotType, req.PlanID,
		req.InstanceLabel, req.DurationDays, req.IsDedicated, req.IsZeroSetup,
	)
	if err != nil {
		h.logger.Error("admin grant failed", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":      "Subscription granted successfully",
		"subscription": sub,
	})
}

func (h *HTTPHandler) CreatePromo(w http.ResponseWriter, r *http.Request) {
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

	var req CreatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	promo := &domain.PromoCode{
		ID:            fmt.Sprintf("promo-%s", uuid.New().String()[:8]),
		Code:          domain.NormalizeCode(req.Code),
		DiscountType:  domain.DiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		MaxUses:       req.MaxUses,
		CurrentUses:   0,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
	}

	if req.DurationDays > 0 {
		exp := time.Now().UTC().Add(time.Duration(req.DurationDays) * 24 * time.Hour)
		promo.ExpiresAt = &exp
	}

	if err := h.service.CreatePromoCode(r.Context(), promo); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"message": "Promo code created",
		"promo":   promo,
	})
}

func (h *HTTPHandler) CreateVoucher(w http.ResponseWriter, r *http.Request) {
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

	var req CreateVoucherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	code := req.Code
	if code == "" {
		code = fmt.Sprintf("GIFT-%s", uuid.New().String()[:8])
	}

	voucher := &domain.VoucherCode{
		ID:           fmt.Sprintf("vouch-%s", uuid.New().String()[:8]),
		Code:         domain.NormalizeVoucher(code),
		PlanID:       req.PlanID,
		BotType:      req.BotType,
		DurationDays: req.DurationDays,
		IsDedicated:  req.IsDedicated,
		IsRedeemed:   false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.service.CreateVoucher(r.Context(), voucher); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"message": "Gift voucher created",
		"voucher": voucher,
	})
}

func (h *HTTPHandler) GetGuildSubscription(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("guildId")
	subs, err := h.service.GetGuildSubscriptions(r.Context(), guildID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve subscriptions")
		return
	}

	var firstSub *domain.Subscription
	for i := range subs {
		if subs[i].IsActiveNow() {
			firstSub = &subs[i]
			break
		}
	}
	if firstSub == nil && len(subs) > 0 {
		firstSub = &subs[0]
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"subscriptions": subs,
		"subscription":  firstSub,
	})
}

func (h *HTTPHandler) HandlePayPalWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Failed reading webhook body")
		return
	}

	var evt struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(body, &evt); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid webhook JSON")
		return
	}

	if err := h.service.HandlePayPalWebhook(r.Context(), evt.EventType, body); err != nil {
		h.logger.Error("error processing paypal webhook", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "Error handling webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
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
