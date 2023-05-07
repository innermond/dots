package postgres_test

import (
	"os"
	"testing"

	"github.com/innermond/dots/postgres"
	"github.com/joho/godotenv"
)

func TestDB(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}

	dsn := os.Getenv("DOTS_DSN")
	db := MustOpenDB(t, dsn)
	defer MustCloseDB(t, db)
}

func MustOpenDB(t *testing.T, dsn string) *postgres.DB {
	t.Helper()

	db := postgres.NewDB(dsn)
	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func MustCloseDB(t *testing.T, db *postgres.DB) {
	t.Helper()

	err := db.Close()
	if err != nil {
		t.Fatal(err)
	}
}
