package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidBotName     = errors.New("bot name cannot be empty")
	ErrInvalidBotCategory = errors.New("invalid bot category")
)

type BotCategory string

const (
	CategoryMusic     BotCategory = "music"
	CategoryGame      BotCategory = "game"
	CategoryUtility   BotCategory = "utility"
	CategoryModerator BotCategory = "moderation"
)

// BotTemplate defines a bot offering available in the store.
type BotTemplate struct {
	ID                 string      `json:"id"`
	Slug               string      `json:"slug"` // URL-friendly identifier, e.g. "groove-music"
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	Category           BotCategory `json:"category"`
	DockerImage        string      `json:"docker_image"`
	DefaultImageTag    string      `json:"default_image_tag"`
	SupportsDedicated  bool        `json:"supports_dedicated"` // Can be deployed as single-tenant
	SupportsShared     bool        `json:"supports_shared"`    // Can be used on shared cluster
	Plans              []Plan      `json:"plans,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// Validate checks if the bot template conforms to business rules.
func (b *BotTemplate) Validate() error {
	if b.Name == "" {
		return ErrInvalidBotName
	}
	switch b.Category {
	case CategoryMusic, CategoryGame, CategoryUtility, CategoryModerator:
		return nil
	default:
		return ErrInvalidBotCategory
	}
}
