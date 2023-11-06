package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
)

type EntryTypeService struct {
	db *DB
}

func NewEntryTypeService(db *DB) *EntryTypeService {
	return &EntryTypeService{db: db}
}

func (s *EntryTypeService) CreateEntryType(ctx context.Context, et *dots.EntryType) error {
	if err := et.Validate(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if canerr := dots.CanCreateOwn(ctx); canerr != nil {
		return canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return err
	}

	if err := createEntryType(ctx, tx, et); err != nil {
		return perr(err)
	}

	tx.Commit()

	return nil
}

func (s *EntryTypeService) FindEntryType(ctx context.Context, filter dots.EntryTypeFilter) ([]*dots.EntryType, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return nil, 0, err
	}

	return findEntryType(ctx, tx, filter)
}

func (s *EntryTypeService) FindEntryTypeUnit(ctx context.Context) ([]string, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return nil, 0, err
	}

	return findEntryTypeUnit(ctx, tx)
}

func (s *EntryTypeService) UpdateEntryType(ctx context.Context, id int, upd dots.EntryTypeUpdate) (*dots.EntryType, error) {
	if err := upd.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return nil, err
	}

	isDeleted := false
	find := dots.EntryTypeFilter{
		ID:        &id,
		IsDeleted: &isDeleted,
	}
	_, n, err := s.FindEntryType(ctx, find)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "entry type not found")
	}

	if canerr := dots.CanWriteOwn(ctx); canerr != nil {
		return nil, canerr
	}

	et, err := updateEntryType(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return et, nil
}

func (s *EntryTypeService) DeleteEntryType(ctx context.Context, id int, filter dots.EntryTypeDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDeleteOwn(ctx); canerr != nil {
		return 0, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return 0, err
	}

	var n int

	if filter.Hard {
		n, err = deleteEntryTypePermanently(ctx, tx, id)
	} else {
		n, err = deleteEntryType(ctx, tx, id, filter.Resurect)
	}
	if err != nil {
		return n, err
	}

	tx.Commit()

	return n, err
}

func createEntryType(ctx context.Context, tx *Tx, et *dots.EntryType) error {
	sqlstr, args := `
insert into entry_type
(code, unit, description)
values
($1, $2, $3) returning id
`, []interface{}{et.Code, et.Unit, et.Description}

	if err := tx.QueryRowContext(
		ctx,
		sqlstr,
		args...,
	).Scan(&et.ID); err != nil {
		return err
	}

	return nil
}

func updateEntryType(ctx context.Context, tx *Tx, id int, updata dots.EntryTypeUpdate) (*dots.EntryType, error) {
	ee, _, err := findEntryType(ctx, tx, dots.EntryTypeFilter{ID: &id, Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("postgres.entry type: cannot retrieve entry type %w", err)
	}
	if len(ee) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "entry type not found")
	}
	et := ee[0]

	set, args := []string{}, []interface{}{}
	if v := updata.Code; v != nil {
		et.Code = v
		set, args = append(set, "code = ?"), append(args, *v)
	}
	if v := updata.Unit; v != nil {
		et.Unit = v
		set, args = append(set, "unit = ?"), append(args, *v)
	}
	if v := updata.Description; v != nil {
		et.Description = v
		set, args = append(set, "description = ?"), append(args, *v)
	}
	replaceQuestionMark(set, args)
	args = append(args, id)

	sqlstr := `
		update entry_type
		set ` + strings.Join(set, ", ") + `
		where	id = ` + fmt.Sprintf("$%d", len(args))

	_, err = tx.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.entry type: cannot update %w", err)
	}

	return et, nil
}

func findEntryType(ctx context.Context, tx *Tx, filter dots.EntryTypeFilter) (_ []*dots.EntryType, n int, err error) {
	where, args := []string{}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Code; v != nil {
		where, args = append(where, "code = ?"), append(args, *v)
	}
	if v := filter.Unit; v != nil {
		where, args = append(where, "unit = ?"), append(args, *v)
	}

	/*v := filter.IsDeleted
	if v != nil && *v {
		where = append(where, "deleted_at is not null")
	} else if v != nil && !*v {
		where = append(where, "deleted_at is null")
	} else if v == nil {
		where = append(where, "deleted_at is null")
	}*/

	wherestr := ""
	if len(where) > 0 {
		replaceQuestionMark(where, args)
		wherestr = "where " + strings.Join(where, " and ")
	}
	sqlstr := `select id, code, description, unit, count(*) over() from entry_type
	` + wherestr + ` ` + formatLimitOffset(filter.Limit, filter.Offset)
	rows, err := tx.QueryContext(ctx,
		sqlstr,
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
		err := rows.Scan(&et.ID, &et.Code, &et.Description, &et.Unit, &n)
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

func findEntryTypeUnit(ctx context.Context, tx *Tx) (_ []string, n int, err error) {
	/*v := filter.IsDeleted
	if v != nil && *v {
		where = append(where, "deleted_at is not null")
	} else if v != nil && !*v {
		where = append(where, "deleted_at is null")
	} else if v == nil {
		where = append(where, "deleted_at is null")
	}*/

	sqlstr := "select entry_type.unit from entry_type group by unit"
	rows, err := tx.QueryContext(ctx, sqlstr)
	if err == sql.ErrNoRows {
		return nil, 0, dots.Errorf(dots.ENOTFOUND, "entry type unit not found")
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entryTypeUnits := []string{}
	for rows.Next() {
		var u string
		err := rows.Scan(&u)
		if err != nil {
			return nil, 0, err
		}
		entryTypeUnits = append(entryTypeUnits, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	n = len(entryTypeUnits)

	return entryTypeUnits, n, nil
}

func deleteEntryType(ctx context.Context, tx *Tx, id int, resurect bool) (n int, err error) {
	where := []string{"core.entry_type.id = $1"}

	kind := "date_trunc('minute', now())::timestamptz"
	if resurect {
		kind = "null"
		where = append(where, "core.entry_type.deleted_at is not null")
	} else {
		where = append(where, "core.entry_type.deleted_at is null")
	}

	wherestr := "where " + strings.Join(where, " and ")

	bareEntryType := `
		and not exists(
		select et.id from entry_type et join entry e on(et.id = e.entry_type_id)
		where e.entry_type_id = $1 limit 1)
`
	sqlstr := `update core.entry_type set deleted_at = %s  ` + wherestr + bareEntryType
	sqlstr = fmt.Sprintf(sqlstr, kind)

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.entry type: cannot soft delete %w", err)
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func deleteEntryTypePermanently(ctx context.Context, tx *Tx, id int) (n int, err error) {
	where := []string{"core.entry_type.id = $1"}
	wherestr := "where " + strings.Join(where, " and ")

	bareEntryType := `
		and not exists(
		select et.id from entry_type et join entry e on(et.id = e.entry_type_id)
		where e.entry_type_id = $1 limit 1)
`
	sqlstr := `delete from core.entry_type ` + wherestr + bareEntryType

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.entry type: cannot soft delete %w", err)
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func entryTypeBelongsToUser(ctx context.Context, tx *Tx, u ksuid.KSUID, e int) error {
	sqlstr := `select exists(select e.id
from entry_type e
where e.tid = $1 and e.id = $2);
`
	var exists bool
	err := tx.QueryRowContext(ctx, sqlstr, u, e).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return dots.Errorf(dots.EUNAUTHORIZED, "foreign entry")
	}

	return nil
}
