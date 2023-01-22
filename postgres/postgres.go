package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx"
)

type DB struct {
	db  *sql.DB
	ctx context.Context
	DSN string
}
