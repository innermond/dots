package postgres

import (
	"context"
	"fmt"

	"github.com/innermond/dots"
)

type DrainService struct {
	db *DB
}

func NewDrainService(db *DB) *DrainService {
	return &DrainService{db: db}
}

func (s *DrainService) CreateOrUpdateDrain(ctx context.Context, d dots.Drain) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return createOrUpdateDrain(ctx, tx, d)
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

	if err := createOrUpdateDrain(ctx, tx, d); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func createOrUpdateDrain(ctx context.Context, tx *Tx, d dots.Drain) error {
	if err := d.Validate(); err != nil {
		return err
	}

  sqlstr := `
insert into drain
(deed_id, entry_id, quantity, is_deleted)
values
($1, $2, $3, $4)
on conflict (deed_id, entry_id) do update set quantity = EXCLUDED.quantity
		`
	_, err := tx.ExecContext(
		ctx,
    sqlstr,
		d.DeedID, d.EntryID, d.Quantity, d.IsDeleted,
	)
  fmt.Println(sqlstr, d)
	if err != nil {
		return err
	}

	return nil
}

func deleteDrainsOfDeed(ctx context.Context, tx *Tx, id int) error {
	_, err := tx.ExecContext(
		ctx,
		`
update drain set is_deleted = true where deed_id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}
