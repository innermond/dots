package postgres

import (
	"context"

	"github.com/innermond/dots"
)

type EntryService struct {
	db *DB
}

func NewEntryService(db *DB) *EntryService {
	return &EntryService{db: db}
}

func (s *EntryService) CreateEntry(ctx context.Context, e *dots.Entry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createEntry(ctx, tx, e); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func createEntry(ctx context.Context, tx *Tx, e *dots.Entry) error {
	user := dots.UserFromContext(ctx)
	if user.ID == 0 {
		return dots.Errorf(dots.EUNAUTHORIZED, "unauthorized user")
	}

	if err := e.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into entry
(entry_type_id, quantity, company_id)
values
($1, $2, $3) returning id, date_added
		`,
		e.EntryTypeID, e.Quantity, e.CompanyID,
	).Scan(&e.ID, &e.DateAdded)
	if err != nil {
		return err
	}

	return nil
}
