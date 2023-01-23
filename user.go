package dots

import (
	"context"
	"time"
)

type User struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Created_on time.Time `json:"created_on"`
	LastLogin  time.Time `json:"last_login"`
}

var UserZero = &User{}

func UserIsZero(u *User) bool {
	return u.ID == 0 && u.Name == "" && u.Created_on.IsZero() && u.LastLogin.IsZero()
}

type UserService interface {
	CreateUser(ctx context.Context, u *User) error
}
