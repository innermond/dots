package dots

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/innermond/dots/autz"
)

type User struct {
	ID     int          `json:"id"`
	Name   string       `json:"name"`
	Email  string       `json:"email"`
	ApiKey string       `json:"api_key"`
	Power  []autz.Power `json:"power"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Auths []*Auth `json:"auths"`
}

type UserFilter struct {
	ID     *int    `json:"id"`
	Email  *string `json:"email"`
	ApiKey *string `json:"api_key"`

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
	return u.ID == 0 &&
		u.Name == "" &&
		u.Email == "" &&
		u.ApiKey == "" &&
		u.Power == nil &&
		u.CreatedAt.IsZero() &&
		u.UpdatedAt.IsZero()
}

type UserService interface {
	CreateUser(context.Context, *User) error
	FindUser(context.Context, UserFilter) ([]*User, int, error)
	FindUserByID(context.Context, int) (*User, error)
}

func (u User) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *User) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("dots.user type assertion failed")
	}
	return json.Unmarshal(b, &u)
}

type userAlias User
type userTimeString struct {
	userAlias

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (u *User) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		return nil
	}

	var user userTimeString
	if err := json.Unmarshal(b, &user); err != nil {
		return err
	}
	layout := "2006-01-02T15:04:05+00:00"
	createdAt, err := time.Parse(layout, user.CreatedAt)
	if err != nil {
		return err
	}
	updatedAt, err := time.Parse(layout, user.UpdatedAt)
	if err != nil {
		return err
	}

	u.ID = user.ID
	u.Name = user.Name
	u.Email = user.Email
	u.ApiKey = user.ApiKey
	u.Power = user.Power
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt

	return nil
}
