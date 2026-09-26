package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidPlanPrice    = errors.New("plan price must be non-negative")
	ErrInvalidPlanInterval = errors.New("plan interval must be monthly or yearly")
)

type BillingInterval string

const (
	IntervalMonthly BillingInterval = "monthly"
	IntervalYearly  BillingInterval = "yearly"
)

// Plan defines a subscription pricing tier for a bot template.
type Plan struct {
	ID          string          `json:"id"`
	BotID       string          `json:"bot_id"`
	Name        string          `json:"name"`        // e.g. "Pro Dedicated", "Free Shared"
	Description string          `json:"description"`
	Interval    BillingInterval `json:"interval"`
	PriceCents  int64           `json:"price_cents"` // In USD cents (e.g. 999 = $9.99)
	Currency    string          `json:"currency"`    // e.g. "USD"
	IsDedicated bool            `json:"is_dedicated"`
	Features    []string        `json:"features"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Validate checks plan invariant rules.
func (p *Plan) Validate() error {
	if p.PriceCents < 0 {
		return ErrInvalidPlanPrice
	}
	if p.Interval != IntervalMonthly && p.Interval != IntervalYearly {
		return ErrInvalidPlanInterval
	}
	return nil
}
