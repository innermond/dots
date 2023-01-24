package postgres

import (
	"context"

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

	result, err := tx.ExecContext(
		ctx,
		`insert into user 
		(name, created_on) values
		(?, ?)`,
		u.Name, u.CreatedOn,
	)
	if err != nil {
		return err
	}

	uid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(uid)

	return nil
}
