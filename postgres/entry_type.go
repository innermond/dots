package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (s *EntryTypeService) UpdateEntryType(ctx context.Context, id int, upd *dots.EntryTypeUpdate) (*dots.EntryType, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	et, err := updateEntryType(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return et, nil

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

func updateEntryType(ctx context.Context, tx *Tx, id int, upd *dots.EntryTypeUpdate) (*dots.EntryType, error) {
	return nil, nil
}

func findEntryType(ctx context.Context, tx *Tx, filter dots.EntryTypeFilter) (_ []*dots.EntryType, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Code; v != nil {
		where, args = append(where, "code = ?"), append(args, *v)
	}
	if v := filter.Unit; v != nil {
		where, args = append(where, "unit = ?"), append(args, *v)
	}
	if v := filter.Tid; v != nil {
		where, args = append(where, "tid = ?"), append(args, *v)
	}
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}

	rows, err := tx.QueryContext(ctx, `
		select id, code, description, unit, tid, count(*) over() from entry_code
		where `+strings.Join(where, " and ")+` `+formatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err == sql.ErrNoRows {
		return nil, 0, dots.Errorf(dots.ENOTFOUND, "entry type not found")
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entryTypes := []*dots.EntryType{}
	for rows.Next() {
		var et dots.EntryType
		err := rows.Scan(&et.ID, &et.Code, &et.Unit, &et.Tid, &n)
		if err != nil {
			return nil, 0, err
		}
		entryTypes = append(entryTypes, &et)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return entryTypes, n, nil
}
