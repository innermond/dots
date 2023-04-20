package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/innermond/dots"
)

type DeedService struct {
	db *DB
}

func NewDeedService(db *DB) *DeedService {
	return &DeedService{db: db}
}

func (s *DeedService) CreateDeed(ctx context.Context, d *dots.Deed) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return createDeed(ctx, tx, d)
	}

	if canerr := dots.CanCreateOwn(ctx); canerr != nil {
		return canerr
	}

	uid := dots.UserFromContext(ctx).ID

	if err := companyBelongsToUser(ctx, tx, uid, d.CompanyID); err != nil {
		return err
	}

	if d.EntryID != nil && d.DrainedQuantity != nil {
		// lock create to own
		// need deed ID and entry ID that belong to companies of user
		err = entryBelongsToUser(ctx, tx, uid, *d.EntryID)
		if err != nil {
			return err
		}
	}

	if err := createDeed(ctx, tx, d); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (s *DeedService) FindDeed(ctx context.Context, filter dots.DeedFilter) ([]*dots.Deed, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return findDeed(ctx, tx, filter, nil)
	}

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	// check search to own
	uid := dots.UserFromContext(ctx).ID
	if filter.CompanyID != nil {
		err := companyBelongsToUser(ctx, tx, uid, *filter.CompanyID)
		if err != nil {
			return nil, 0, err
		}
		return findDeed(ctx, tx, filter, nil)
	} else {
		// lock search to own
		return findDeed(ctx, tx, filter, &uid)
	}
}

func (s *DeedService) UpdateDeed(ctx context.Context, id int, upd dots.DeedUpdate) (*dots.Deed, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return updateDeed(ctx, tx, id, upd)
	}

	uid := dots.UserFromContext(ctx).ID

	if upd.CompanyID != nil {
		err = companyBelongsToUser(ctx, tx, uid, *upd.CompanyID)
		if err != nil {
			return nil, err
		}
	}

	if err := deedBelongsToUser(ctx, tx, uid, id); err != nil {
		return nil, err
	}

	deedUserID := deedGetUser(ctx, tx, id)
	if deedUserID == nil {
		return nil, dots.Errorf(dots.EINVALID, "deed user conflict")
	}
	if canerr := dots.CanWriteOwn(ctx, *deedUserID); canerr != nil {
		return nil, canerr
	}

	d, err := updateDeed(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return d, nil
}

func (s *DeedService) DeleteDeed(ctx context.Context, filter dots.DeedDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return deleteDeed(ctx, tx, filter, nil)
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
		n, err = deleteDeed(ctx, tx, filter, nil)
	} else {
		// lock delete to own
		n, err = deleteDeed(ctx, tx, filter, &uid)
	}

	tx.Commit()

	return n, err
}

func createDeed(ctx context.Context, tx *Tx, d *dots.Deed) error {
	user := dots.UserFromContext(ctx)
	if user.ID == 0 {
		return dots.Errorf(dots.EUNAUTHORIZED, "unauthorized user")
	}

	if err := d.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into deed
(title, quantity, unit, unitprice, company_id)
values
($1, $2, $3, $4, $5) returning id
		`,
		d.Title, d.Quantity, d.Unit, d.UnitPrice, d.CompanyID,
	).Scan(&d.ID)
	if err != nil {
		return err
	}

	if d.EntryID != nil && d.DrainedQuantity != nil {
		d := dots.Drain{
			DeedID:   d.ID,
			EntryID:  *d.EntryID,
			Quantity: *d.DrainedQuantity,
		}

		err = createOrUpdateDrain(ctx, tx, d)
		if err != nil {
			return err
		}

	}

	return nil
}

func updateDeed(ctx context.Context, tx *Tx, id int, updata dots.DeedUpdate) (*dots.Deed, error) {
	dd, _, err := findDeed(ctx, tx, dots.DeedFilter{ID: &id, Limit: 1}, nil)
	if err != nil {
		return nil, fmt.Errorf("postgres.deed: cannot retrieve deed %w", err)
	}
	if len(dd) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "deed not found")
	}
	e := dd[0]

	set, args := []string{}, []interface{}{}
	if v := updata.Title; v != nil {
		e.Title = *v
		set, args = append(set, "title = ?"), append(args, *v)
	}
	if v := updata.Quantity; v != nil {
		e.Quantity = *v
		set, args = append(set, "quantity = ?"), append(args, *v)
	}
	if v := updata.Unit; v != nil {
		e.Unit = *v
		set, args = append(set, "unit = ?"), append(args, *v)
	}
	if v := updata.UnitPrice; v != nil {
		e.UnitPrice = *v
		set, args = append(set, "unitprice = ?"), append(args, *v)
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
		update deed
		set ` + strings.Join(set, ", ") + `
		where	id = ` + fmt.Sprintf("$%d", len(args))

	_, err = tx.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.deed: cannot update %w", err)
	}

	if updata.EntryID != nil && updata.DrainedQuantity != nil {
		d := dots.Drain{
			DeedID:   id,
			EntryID:  *updata.EntryID,
			Quantity: *updata.DrainedQuantity,
		}

		err = createOrUpdateDrain(ctx, tx, d)
		if err != nil {
			return e, err
		}

	}

	return e, nil
}

func findDeed(ctx context.Context, tx *Tx, filter dots.DeedFilter, lockOwnID *int) (_ []*dots.Deed, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Title; v != nil {
		where, args = append(where, "title = ?"), append(args, *v)
	}
	if v := filter.Quantity; v != nil {
		where, args = append(where, "quantity = ?"), append(args, *v)
	}
	if v := filter.Unit; v != nil {
		where, args = append(where, "unit = ?"), append(args, *v)
	}
	if v := filter.UnitPrice; v != nil {
		where, args = append(where, "unitprice = ?"), append(args, *v)
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

	// WARN: placeholder ? is connected with position in "where"
	// so any unrelated with position (read replacement $n)
	// MUST be added AFTER the "for" cycle
	// that binds value with placeholder

	// the presence of deleted key with empty value
	// signals to find ONLY deleted records
	if filter.DeletedAtTo == nil && filter.DeletedAtFrom == nil {
		where = append(where, "deleted_at is null")
	}

	sqlstr := `select id, title, unit, unitprice, quantity, company_id, count(*) over() from deed
		where `
	sqlstr = sqlstr + strings.Join(where, " and ") + ` ` + formatLimitOffset(filter.Limit, filter.Offset)
	rows, err := tx.QueryContext(
		ctx,
		sqlstr,
		args...,
	)
	if err == sql.ErrNoRows {
		return nil, 0, dots.Errorf(dots.ENOTFOUND, "deed not found")
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	deeds := []*dots.Deed{}
	for rows.Next() {
		var d dots.Deed
		err := rows.Scan(&d.ID, &d.Title, &d.Unit, &d.UnitPrice, &d.Quantity, &d.CompanyID, &n)
		if err != nil {
			return nil, 0, err
		}
		deeds = append(deeds, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return deeds, n, nil
}

func deleteDeed(ctx context.Context, tx *Tx, filter dots.DeedDelete, lockOwnID *int) (n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Title; v != nil {
		where, args = append(where, "title = ?"), append(args, *v)
	}
	if v := filter.Quantity; v != nil {
		where, args = append(where, "quantity = ?"), append(args, *v)
	}
	if v := filter.Unit; v != nil {
		where, args = append(where, "unit = ?"), append(args, *v)
	}
	if v := filter.UnitPrice; v != nil {
		where, args = append(where, "unitprice = ?"), append(args, *v)
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

	kind := "date_trunc('second;, now())::timestamptz"
	if filter.Resurect {
		kind = "null"
		where = append(where, "deleted_at is not null")
	} else {
		where = append(where, "deleted_at is null")
	}
	sqlstr := "update deed set deleted_at = " + kind + " where "
	sqlstr = sqlstr + strings.Join(where, " and ") + " " + formatLimitOffset(filter.Limit, filter.Offset)

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.deed: cannot soft delete %w", err)
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func deedBelongsToUser(ctx context.Context, tx *Tx, u int, d int) error {
	sqlstr := `select exists(select d.id
from deed d
where d.company_id = any(select id
from company c
where c.tid = $1)
and d.id = $2);
`
	var exists bool
	err := tx.QueryRowContext(ctx, sqlstr, u, d).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return dots.Errorf(dots.EUNAUTHORIZED, "foreign deed")
	}

	return nil
}

func deedGetUser(ctx context.Context, tx *Tx, d int) *int {
	sqlstr := `select c.tid
from company c
where c.id = (select d.company_id 
from deed d
where d.id = $1)
`
	var uid int
	err := tx.QueryRowContext(ctx, sqlstr, d).Scan(&uid)
	if err != nil {
		return nil
	}

	return &uid
}
