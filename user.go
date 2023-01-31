package dots

import (
	"context"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UserFilter struct {
	ID    *int    `json:"id"`
	Email *string `json:"email"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
type UserUpdate struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (u *User) Validate() error {
	// TODO regex for detecting white spaces
	if u.Name == "" {
		return Errorf(ECONFLICT, "User name required.")
	}
	return nil
}

var UserZero = &User{}

func UserIsZero(u *User) bool {
	return u.ID == 0 && u.Name == "" && u.CreatedAt.IsZero() && u.UpdatedAt.IsZero()
}

type UserService interface {
	CreateUser(context.Context, *User) error
	FindUser(context.Context, UserFilter) ([]*User, int, error)
}
