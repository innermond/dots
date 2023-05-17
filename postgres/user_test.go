package postgres_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/innermond/dots"
	"github.com/innermond/dots/postgres"
	"github.com/segmentio/ksuid"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		u0 := &dots.User{
			Name:      "U0",
			Email:     "USER_EMAIL@FOO.COM",
			CreatedAt: db.Now().Add(-24 * time.Hour),
		}

		ctx0, deleteUser0 := MustCreateUser(t, context.Background(), db, u0)
		defer deleteUser0()

		if u0.ID == ksuid.Nil {
			t.Fail()
			t.Logf("user ID expected to be different than %v", ksuid.Nil)
		}

		if u0.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if u0.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		t.Logf("user created: %+v", u0)

		s := postgres.NewUserService(db)
		if other, err := s.FindUserByID(ctx0, u0.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(u0, other) {
			t.Fatalf("mismatch: %#v != %#v", u0, other)
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		s := postgres.NewUserService(db)
		u0 := &dots.User{Name: "ORIGINAL NAME"}
		ctx0, deleteU0 := MustCreateUser(t, context.Background(), db, u0)
		defer deleteU0()

		newName := "UPDATED NAME"
		updated, err := s.UpdateUser(ctx0, u0.ID, dots.UserUpdate{Name: &newName})
		if err != nil {
			t.Fatal(err)
		}

		got, want := updated.Name, newName
		if got != want {
			t.Fatalf("Name=%v, want %v", got, want)
		}

		other, err := s.FindUserByID(context.Background(), u0.ID)
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(updated, other) {
			t.Fatalf("mismatch: %#v != %#v", updated, other)
		}
	})
}

func MustCreateUser(t *testing.T, ctx context.Context, db *postgres.DB, user *dots.User) (context.Context, func()) {
	t.Helper()
	s := postgres.NewUserService(db)
	err := s.CreateUser(ctx, user)
	if err != nil {
		t.Fatal(err)
	}

	deleteTestUserFunc := func() {
		// delete testing user
		if err := s.DeleteUser(ctx, user.ID); err != nil {
			t.Fatal(err)
		}
		t.Logf("user deleted: %+v", user)
	}

	return dots.NewContextWithUser(ctx, user), deleteTestUserFunc
}
