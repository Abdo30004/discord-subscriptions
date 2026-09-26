package domain_test

import (
	"testing"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
)

func TestBotTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bot     domain.BotTemplate
		wantErr bool
	}{
		{
			name: "valid music bot",
			bot: domain.BotTemplate{
				Name:     "Music Master",
				Category: domain.CategoryMusic,
			},
			wantErr: false,
		},
		{
			name: "empty name fails",
			bot: domain.BotTemplate{
				Name:     "",
				Category: domain.CategoryMusic,
			},
			wantErr: true,
		},
		{
			name: "invalid category fails",
			bot: domain.BotTemplate{
				Name:     "Spam Bot",
				Category: "invalid_category",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestPlan_Validate(t *testing.T) {
	tests := []struct {
		name    string
		plan    domain.Plan
		wantErr bool
	}{
		{
			name: "valid pro plan",
			plan: domain.Plan{
				PriceCents: 999,
				Interval:   domain.IntervalMonthly,
			},
			wantErr: false,
		},
		{
			name: "negative price fails",
			plan: domain.Plan{
				PriceCents: -100,
				Interval:   domain.IntervalMonthly,
			},
			wantErr: true,
		},
		{
			name: "invalid interval fails",
			plan: domain.Plan{
				PriceCents: 500,
				Interval:   "weekly",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}
