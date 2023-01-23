package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/innermond/dots"
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

type PingService struct {
	db *DB
}

func NewPingService(db *DB) *PingService {
	return &PingService{db: db}
}

func (s *PingService) ById(ctx context.Context) *dots.Ping {
	ping := dots.Ping{}
	err := s.db.db.QueryRow("select 111").Scan(&ping.ID)
	if err != nil {
		panic(fmt.Errorf("db error : %w", err))
	}
	return &ping

}
