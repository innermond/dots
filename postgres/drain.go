package postgres

import (
	"context"

	"github.com/innermond/dots"
)

type DrainService struct {
	db *DB
}

func NewDrainService(db *DB) *DrainService {
	return &DrainService{db: db}
}

func (s *DrainService) CreateDrain(ctx context.Context, d *dots.Drain) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createDrain(ctx, tx, d); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func createDrain(ctx context.Context, tx *Tx, d *dots.Drain) error {
	if err := d.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into drain
(deed_id, entry_id, quantity)
values
($1, $2, $3)
		`,
		d.DeedID, d.EntryID, d.Quantity,
	).Scan()
	if err != nil {
		return err
	}

	return nil
}
