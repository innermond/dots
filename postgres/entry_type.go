package postgres

import (
	"context"

	"github.com/innermond/dots"
)

type EntryTypeService struct {
	db *DB
}

func NewEntryTypeService(db *DB) *EntryTypeService {
	return &EntryTypeService{db: db}
}

func (s *EntryTypeService) CreateEntryType(ctx context.Context, et *dots.EntryType) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createEntryType(ctx, tx, et); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func createEntryType(ctx context.Context, tx *Tx, et *dots.EntryType) error {
	user := dots.UserFromContext(ctx)
	if user.ID == 0 {
		return dots.Errorf(dots.EUNAUTHORIZED, "unauthorized user")
	}

	if err := et.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into entry_type
(code, unit, tid)
values
($1, $2, $3) returning id
		`,
		et.Code, et.Unit, user.ID,
	).Scan(&et.ID)
	if err != nil {
		return err
	}
	et.Tid = user.ID

	return nil
}
