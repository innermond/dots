package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	db  *sql.DB
	DSN string

	ctx    context.Context
	cancel func()

	Now func() time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
		Now: time.Now,
	}
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
	db  *DB
	now time.Time
}

func (tx *Tx) RollbackOrCommit(err error) {
	switch err {
	case nil:
		tx.Commit()
	default:
		tx.Rollback()
	}
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		db:  db,
		now: db.Now().UTC().Truncate(time.Second),
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

func timeRFC3339(val sql.NullTime) time.Time {
	if val.Valid {
		loc, err := time.LoadLocation("UTC")
		if err != nil {
			return (*sql.NullTime)(nil).Time
		}
		vs := val.Time.In(loc).Format(time.RFC3339)
		v, err := time.Parse(time.RFC3339, vs)
		if err != nil {
			return (*sql.NullTime)(nil).Time
		}
		return v
	}
	return (*sql.NullTime)(nil).Time
}

// where and args must be constructed togheter to be sync'ed
func replaceQuestionMark(where []string, args []interface{}) {
  // PostgreSQL uses numbered placeholders starting from $1
  for i, v := range where {
    if !strings.Contains(v, "?") {
      continue
    }
    v = strings.Replace(v, "?", fmt.Sprintf("$%d", i+1), 1)
    where[i] = v
  }
}
