package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/innermond/dots"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(dsn string) (pool *pgxpool.Pool, err error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return pool, err
	}

	config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		u := dots.UserFromContext(ctx)
		if u.ID.IsNil() {
			log.Println("user expected to be found")
			return false
		}
		memberId := u.ID.String()
		_, err = conn.Exec(ctx, "SELECT set_tenant($1)", memberId)

		if err != nil {
			log.Fatal(err)
			return false
		} else {
			fmt.Println("Set session to memberId: " + memberId)
		}

		return true
	}

	config.AfterRelease = func(conn *pgx.Conn) bool {
		// set the setting to be empty before this connection is released to pool
		_, err := conn.Exec(context.Background(), "select set_tenant($1)", "")

		if err != nil {
			log.Fatal(err)
			return false
		} else {
			fmt.Println("Cleared the member id")
		}

		return true
	}

	config.MaxConns = int32(20)
	config.MaxConnLifetime = time.Minute
	config.MaxConnIdleTime = time.Minute

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	return pool, err
}
