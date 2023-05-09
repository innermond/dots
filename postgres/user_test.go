package postgres_test

import (
	"context"
	"testing"

	"github.com/innermond/dots"
	"github.com/innermond/dots/postgres"
)

func TestUserService_CreateUser(t *testing.T) {
  t.Run("New User", func(t *testing.T) {
    db := MustOpenDB(t, DSN)
    defer MustCloseDB(t, db)

    s := postgres.NewUserService(db)
    u := &dots.User{}

    ctx := context.Background()
    if err := s.CreateUser(ctx, u); err != nil {
      t.Fatal(err)
    }

  })
}
