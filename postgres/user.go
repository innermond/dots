package postgres

import (
	"context"
	"time"

	"github.com/innermond/dots"
)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, u *dots.User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = createUser(ctx, tx, u); err != nil {
		return err
	}

	return tx.Commit()
}

func createUser(ctx context.Context, tx *Tx, u *dots.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	created_on := time.Now().UTC().Truncate(time.Second)
	err := tx.QueryRowContext(
		ctx, `
		INSERT INTO "user" (
			name,
			created_on
		)
		values ($1, $2) returning id
	`,
		u.Name, created_on,
	).Scan(&u.ID)
	if err != nil {
		return err
	}

	u.CreatedOn = created_on

	return nil
}
