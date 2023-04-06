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

func (s *DrainService) CreateDrain(ctx context.Context, d dots.Drain) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return createDrain(ctx, tx, d)
	}

	if canerr := dots.CanCreateOwn(ctx); canerr != nil {
		return canerr
	}

	// lock create to own
	// need deed ID and entry ID that belong to companies of user
	uid := dots.UserFromContext(ctx).ID
	err = entryBelongsToUser(ctx, tx, uid, d.EntryID)
	if err != nil {
		return err
	}
	err = deedBelongsToUser(ctx, tx, uid, d.DeedID)
	if err != nil {
		return err
	}

	if err := createDrain(ctx, tx, d); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func createDrain(ctx context.Context, tx *Tx, d dots.Drain) error {
	if err := d.Validate(); err != nil {
		return err
	}

	_, err := tx.ExecContext(
		ctx,
		`
insert into drain
(deed_id, entry_id, quantity)
values
($1, $2, $3)
on conflict (deed_id, entry_id) do update set quantity = drain.quantity + EXCLUDED.quantity
		`,
		d.DeedID, d.EntryID, d.Quantity,
	)
	if err != nil {
		return err
	}

	return nil
}
