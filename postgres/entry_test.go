package postgres_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestEntryService_CreateEntry(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal(err)
	}

	dsn := os.Getenv("DOTS_DSN")
	createEntry := func(t *testing.T) {
		db := MustOpenDB(t, dsn)
		defer MustCloseDB(t, db)

		//s := postgres.NewEntryService(db)
		// posgres.CreateEntry
	}
	t.Run("OK", createEntry)
}
