package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (s *EntryService) FindEntry(ctx context.Context, filter dots.EntryFilter) ([]*dots.Entry, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	return findEntry(ctx, tx, filter)
}

func (s *EntryService) UpdateEntry(ctx context.Context, id int, upd dots.EntryUpdate) (*dots.Entry, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	e, err := updateEntry(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return e, nil
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

func updateEntry(ctx context.Context, tx *Tx, id int, updata dots.EntryUpdate) (*dots.Entry, error) {
	ee, _, err := findEntry(ctx, tx, dots.EntryFilter{ID: &id, Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("postgres.entry: cannot retrieve entry %w", err)
	}
	if len(ee) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "entry not found")
	}
	e := ee[0]

	set, args := []string{}, []interface{}{}
	if v := updata.EntryTypeID; v != nil {
		e.EntryTypeID = *v
		set, args = append(set, "entry_type_id = ?"), append(args, *v)
	}
	if v := updata.DateAdded; v != nil {
		e.DateAdded = *v
		set, args = append(set, "date_added = ?"), append(args, *v)
	}
	if v := updata.Quantity; v != nil {
		e.Quantity = *v
		set, args = append(set, "quantity = ?"), append(args, *v)
	}
	if v := updata.CompanyID; v != nil {
		e.CompanyID = *v
		set, args = append(set, "company_id = ?"), append(args, *v)
	}

	for inx, v := range set {
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx+1), 1)
		set[inx] = v
	}
	args = append(args, id)

	sqlstr := `
		update entry
		set ` + strings.Join(set, ", ") + `
		where	id = ` + fmt.Sprintf("$%d", len(args))

	_, err = tx.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.entry: cannot update %w", err)
	}

	return e, nil
}

func findEntry(ctx context.Context, tx *Tx, filter dots.EntryFilter) (_ []*dots.Entry, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.EntryTypeID; v != nil {
		where, args = append(where, "entry_type_id = ?"), append(args, *v)
	}
	if v := filter.DateAdded; v != nil {
		where, args = append(where, "date_added = ?"), append(args, *v)
	}
	if v := filter.Quantity; v != nil {
		where, args = append(where, "quantity = ?"), append(args, *v)
	}
	if v := filter.CompanyID; v != nil {
		where, args = append(where, "company_id = ?"), append(args, *v)
	}
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}

	rows, err := tx.QueryContext(ctx, `
		select id, entry_type_id, date_added, quantity, company_id, count(*) over() from entry
		where `+strings.Join(where, " and ")+` `+formatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err == sql.ErrNoRows {
		return nil, 0, dots.Errorf(dots.ENOTFOUND, "entry not found")
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := []*dots.Entry{}
	for rows.Next() {
		var e dots.Entry
		err := rows.Scan(&e.ID, &e.EntryTypeID, &e.DateAdded, &e.Quantity, &e.CompanyID, &n)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return entries, n, nil
}
