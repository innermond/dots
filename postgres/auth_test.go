package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/innermond/dots"
	"github.com/innermond/dots/postgres"
)

func TestAuthService_CreateAuth(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		s := postgres.NewAuthService(db)
		a := &dots.Auth{
			Source:      "SOURCE",
			SourceID:    "SOURCE_ID",
			AccessToken: "ACCESS_TOKEN",
			User: &dots.User{
				Name: "User Auth",
			},
		}

		if err := s.CreateAuth(context.Background(), a); err != nil {
			t.Fatal(err)
		}

		t.Logf("auth created %#v", a)

		aa, _, err := s.FindAuths(context.Background(), dots.AuthFilter{ID: &a.ID})
		t.Logf("found auth: %+v\n", aa)
		defer func() {
			if err := s.DeleteAuth(context.Background(), a.ID); err != nil {
				t.Fatal(err)
			}
		}()
		if err != nil {
			t.Fatal(err)
			// TODO: aa[0] != a has time Local
		}

		if !reflect.DeepEqual(aa[0], a) {
			t.Fatalf("mismatch: %#v != %#v", aa[0].CreatedAt, a.CreatedAt)
		}

	})

	t.Run("ErrValidation", func(t *testing.T) {
		db := MustOpenDB(t, DSN)
		defer MustCloseDB(t, db)

		s := postgres.NewAuthService(db)

		err := s.CreateAuth(context.Background(), &dots.Auth{User: &dots.User{Name: "User Source"}})
		if err == nil {
			t.Fatal("expected error")
		} else if dots.ErrorCode(err) != dots.EINVALID {
			t.Fatalf("expected EINVALID got: %v", err)
		}

	})
}
