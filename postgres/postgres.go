package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	db  *sql.DB
	DSN string

	ctx    context.Context
	cancel func()
}

func NewDB(dsn string) *DB {
	db := &DB{DSN: dsn}
	db.ctx, db.cancel = context.WithCancel(context.Background())

	return db
}

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return errors.New("dns required")
	}

	db.db, err = sql.Open("pgx", db.DSN)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	db.cancel()
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

type Tx struct {
	*sql.Tx
	db *DB
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx: tx,
		db: db,
	}, nil
}

func formatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("limit %d offset %d", limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf("limit %d", limit)
	} else if offset > 0 {
		return fmt.Sprintf("offset %d", offset)
	}
	return ""
}

func timeRFC3339(val sql.NullTime) *time.Time {
	if val.Valid {
		v, err := time.Parse(time.RFC3339, val.Time.String())
		if err != nil || v.IsZero() {
			return nil
		}
		return &v
	}
	return nil
}
