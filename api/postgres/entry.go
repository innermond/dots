package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
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

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return createEntry(ctx, tx, e)
	}

	if canerr := dots.CanCreateOwn(ctx); canerr != nil {
		return canerr
	}

	uid := dots.UserFromContext(ctx).ID
	if err := companyBelongsToUser(ctx, tx, uid, e.CompanyID); err != nil {
		return err
	}

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

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return findEntry(ctx, tx, filter)
	}

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	// need company ID that belong to user
	if filter.CompanyID == nil {
		return nil, 0, dots.Errorf(dots.EINVALID, "missing company")
	}

	uid := dots.UserFromContext(ctx).ID
	err = companyBelongsToUser(ctx, tx, uid, *filter.CompanyID)
	if err != nil {
		return nil, 0, err
	}

	return findEntry(ctx, tx, filter)
}

func (s *EntryService) UpdateEntry(ctx context.Context, id int, upd dots.EntryUpdate) (*dots.Entry, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return updateEntry(ctx, tx, id, upd)
	}

	uid := dots.UserFromContext(ctx).ID
	if upd.CompanyID != nil {
		err := companyBelongsToUser(ctx, tx, uid, *upd.CompanyID)
		if err != nil {
			return nil, err
		}
	}

	if upd.EntryTypeID != nil {
		err := entryTypeBelongsToUser(ctx, tx, uid, *upd.EntryTypeID)
		if err != nil {
			return nil, err
		}
	}

	if canerr := dots.CanWriteOwn(ctx, uid); canerr != nil {
		return nil, canerr
	}

	e, err := updateEntry(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return e, nil
}

func (s *EntryService) DeleteEntry(ctx context.Context, filter dots.EntryDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return deleteEntry(ctx, tx, filter, nil)
	}

	if canerr := dots.CanDeleteOwn(ctx); canerr != nil {
		return 0, canerr
	}

	var n int
	// check search to own
	uid := dots.UserFromContext(ctx).ID
	if filter.CompanyID != nil {
		err = companyBelongsToUser(ctx, tx, uid, *filter.CompanyID)
		if err != nil {
			return 0, err
		}
		n, err = deleteEntry(ctx, tx, filter, nil)
	} else {
		// lock delete to own
		n, err = deleteEntry(ctx, tx, filter, &uid)
	}

	tx.Commit()

	return n, err
}

func createEntry(ctx context.Context, tx *Tx, e *dots.Entry) error {
	user := dots.UserFromContext(ctx)
	if user.ID == ksuid.Nil {
		return dots.Errorf(dots.EUNAUTHORIZED, "unauthorized user")
	}

	if err := e.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into entry
(entry_type_id, quantity, company_id, date_added)
values
($1, $2, $3, date_trunc('minute', now())::timestamptz) returning id, date_added
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

	// use inx+1 to avoid invalid $0 placeholder
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
	if v := filter.DeletedAtFrom; v != nil {
		// >= ? is intentional
		where, args = append(where, "deleted_at >= ?"), append(args, *v)
	}
	if v := filter.DeletedAtTo; v != nil {
		// < ? is intentional
		// avoid double counting exact midnight values
		where, args = append(where, "deleted_at < ?"), append(args, *v)
	}
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}
	if filter.DeletedAtTo == nil && filter.DeletedAtFrom == nil {
		where = append(where, "deleted_at is null")
	}

	sqlstr := "select id, entry_type_id, date_added, quantity, company_id, count(*) over() from entry where "
	sqlstr = sqlstr + strings.Join(where, " and ") + " " + formatLimitOffset(filter.Limit, filter.Offset)
	fmt.Println(sqlstr)
	rows, err := tx.QueryContext(
		ctx,
		sqlstr,
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

func deleteEntry(ctx context.Context, tx *Tx, filter dots.EntryDelete, lockOwnID *ksuid.KSUID) (n int, err error) {
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
	if v := filter.DeletedAtFrom; v != nil {
		// >= ? is intentional
		where, args = append(where, "deleted_at >= ?"), append(args, *v)
	}
	if v := filter.DeletedAtTo; v != nil {
		// < ? is intentional
		// avoid double counting exact midnight values
		where, args = append(where, "deleted_at < ?"), append(args, *v)
	}
	if lockOwnID != nil {
		where, args = append(where, "company_id = any(select id from company where tid = ?)"), append(args, *lockOwnID)
	}
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}
	// "delete" only entries that are not used on drain table
	where = append(where, "d.entry_id is null")

	kind := "date_trunc('minute', now())::timestamptz"
	if filter.Resurect {
		kind = "null"
		where = append(where, "e.deleted_at is not null")
	} else {
		where = append(where, "e.deleted_at is null")
	}

	sqlstr := `update entry set deleted_at = %s where id = any(
		select id
		from entry e left join drain d on(e.id = d.entry_id)
		where %s)`
	sqlstr = fmt.Sprintf(sqlstr, kind, strings.Join(where, " and ")+" "+formatLimitOffset(filter.Limit, filter.Offset))

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.entry: cannot soft delete %w", err)
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func entryBelongsToUser(ctx context.Context, tx *Tx, u ksuid.KSUID, e int) error {
	sqlstr := `select exists(select e.id
from entry e
where e.company_id = any(select id
from company c
where c.tid = $1)
and e.id = $2);
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
