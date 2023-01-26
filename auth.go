package dots

import (
	"context"
	"time"
)

type Auth struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Source       string    `json:"source"`
	SourceID     string    `json:"source_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthFilter struct {
	ID        *int       `json:"id"`
	UserID    *int       `json:"user_id"`
	Source    *string    `json:"source"`
	SourceID  *string    `json:"source_id"`
	Expiry    *time.Time `json:"expity"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type AuthService interface {
	CreateAuth(context.Context, *Auth) error
}
