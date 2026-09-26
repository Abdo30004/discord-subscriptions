package domain

import (
	"fmt"
	"strconv"
)

const (
	PermissionAdministrator int64 = 0x8
	PermissionManageGuild   int64 = 0x20
)

// Guild represents a Discord server associated with the authenticated user.
type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon,omitempty"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
	CanManage   bool   `json:"can_manage"`
}

// ComputeCanManage checks if user has Administrator or Manage Guild permissions.
func (g *Guild) ComputeCanManage() bool {
	if g.Owner {
		g.CanManage = true
		return true
	}

	permInt, err := strconv.ParseInt(g.Permissions, 10, 64)
	if err != nil {
		g.CanManage = false
		return false
	}

	isAdmin := (permInt & PermissionAdministrator) == PermissionAdministrator
	isManager := (permInt & PermissionManageGuild) == PermissionManageGuild

	g.CanManage = isAdmin || isManager
	return g.CanManage
}

// IconURL returns the Discord CDN image for the server icon.
func (g *Guild) IconURL() string {
	if g.Icon == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", g.ID, g.Icon)
}
