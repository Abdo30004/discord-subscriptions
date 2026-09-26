package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrVoucherAlreadyRedeemed = errors.New("voucher code has already been redeemed")
	ErrInvalidVoucherCode     = errors.New("invalid voucher code")
)

// VoucherCode represents a one-time gift card / subscription activation code.
type VoucherCode struct {
	ID                 string     `json:"id"`
	Code               string     `json:"code"`
	PlanID             string     `json:"plan_id"`
	BotType            string     `json:"bot_type"`
	DurationDays       int        `json:"duration_days"`
	IsDedicated        bool       `json:"is_dedicated"`
	IsRedeemed         bool       `json:"is_redeemed"`
	RedeemedByUserID   string     `json:"redeemed_by_user_id,omitempty"`
	RedeemedByGuildID  string     `json:"redeemed_by_guild_id,omitempty"`
	RedeemedAt         *time.Time `json:"redeemed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// Redeem marks the voucher as claimed by a user for their guild.
func (v *VoucherCode) Redeem(userID, guildID string) error {
	if v.IsRedeemed {
		return ErrVoucherAlreadyRedeemed
	}
	if userID == "" || guildID == "" {
		return errors.New("user ID and guild ID are required to redeem voucher")
	}

	now := time.Now().UTC()
	v.IsRedeemed = true
	v.RedeemedByUserID = userID
	v.RedeemedByGuildID = guildID
	v.RedeemedAt = &now
	return nil
}

// NormalizeVoucher standardizes voucher format.
func NormalizeVoucher(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
