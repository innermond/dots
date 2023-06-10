package dots

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

const (
	AuthSourceGithub = "github"
	AuthSourceGoogle = "google"
)

type Auth struct {
	ID           int         `json:"id"`
	UserID       ksuid.KSUID `json:"user_id"`
	User         *User       `json:"user"`
	Source       string      `json:"source"`
	SourceID     string      `json:"source_id"`
	AccessToken  string      `json:"-"`
	RefreshToken string      `json:"-"`
	Expiry       *time.Time  `json:"-"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (a *Auth) IsUserPersisted() bool {
	return a.User == nil || a.User.ID == ksuid.Nil
}

func (a *Auth) Validate() error {
	if a.UserID == ksuid.Nil {
		return Errorf(EINVALID, "user required")
	} else if a.Source == "" {
		return Errorf(EINVALID, "source required")
	} else if a.SourceID == "" {
		return Errorf(EINVALID, "source ID required")
	} else if a.AccessToken == "" {
		return Errorf(EINVALID, "access token required")
	}
	return nil
}

type AuthFilter struct {
	ID        *int         `json:"id"`
	UserID    *ksuid.KSUID `json:"user_id"`
	Source    *string      `json:"source"`
	SourceID  *string      `json:"source_id"`
	Expiry    *time.Time   `json:"expiry"`
	CreatedAt *time.Time   `json:"created_at"`
	UpdatedAt *time.Time   `json:"updated_at"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type AuthService interface {
	CreateAuth(context.Context, *Auth) error
}

type TokenService interface {
  Create() (string, error)
}
