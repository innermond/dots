package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
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

	// order for distribute is important
	// try first automatic distribute
	if len(d.EntryTypeDistribute) > 0 {
		etids := getEntryIDsFromDistribute(d.EntryTypeDistribute)
		etqty, err := quantityByEntryTypes(ctx, tx, etids)
		if err != nil {
			return err
		}

		notenough := map[int]float64{}
		for k, wanted := range d.EntryTypeDistribute {
			if existent, found := etqty[k]; !found {
				return dots.Errorf(dots.ENOTFOUND, "not found entry type %v", k)
			} else if wanted > existent {
				notenough[k] = wanted - existent
			}
		}

		if len(notenough) > 0 {
			err := &dots.Error{
				Code:    dots.EINVALID,
				Message: "not enough quantity",
				Data:    map[string]interface{}{"notenough": notenough},
			}
			return err
		}
		// suggestDistributeOverEntryType
		distribute, err := suggestDistributeOverEntryType(ctx, tx, d.EntryTypeDistribute)
		if err != nil {
			return err
		}
		d.Distribute = distribute
	}

	// ensures to have something to process
	if len(d.Distribute) > 0 {
		// check entries are owned and enough
		// this doesn't check user ownership over entries
		check, err := entriesOfCompanyAreEnough(ctx, tx, d.Distribute, d.CompanyID)
		if err != nil {
			if err == sql.ErrNoRows {
				return dots.Errorf(dots.ENOTFOUND, "entries owned and enough not found")
			}
			return err
		}
		// need to check check
		notenough := map[int]float64{}
		for eid, diff := range check {
			if diff < 0 {
				notenough[eid] = diff
			}
		}
		// not enough
		if len(notenough) > 0 {
			err := &dots.Error{
				Code:    dots.ECONFLICT,
				Message: "not enough entries",
				Data:    map[string]interface{}{"notenough": notenough, "company_id": d.CompanyID},
			}
			return err
		}
	}

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

	if d.Distribute != nil {
		ee := getEntryIDsFromDistribute(d.Distribute)
		if len(ee) == 0 {
			err := &dots.Error{
				Code:    dots.EINVALID,
				Message: "entries not specified",
			}
			return err
		}
		// need deed ID and entry ID that belong to companies of user
		err = entriesBelongsToUser(ctx, tx, uid, ee)
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

	// TODO: is CompanyID required for all update operations?
	if len(upd.Distribute) > 0 {
		if upd.CompanyID == nil {
			return nil, &dots.Error{
				Code:    dots.EINVALID,
				Message: "company id is required",
			}
		}

		// check entries are owned and enough
		check, err := entriesOfCompanyAreEnough(ctx, tx, upd.Distribute, *upd.CompanyID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, dots.Errorf(dots.ENOTFOUND, "entries owned and enough not found")
			}
			return nil, err
		}
		// need to check check
		notenough := map[int]float64{}
		for eid, diff := range check {
			if diff < 0 {
				notenough[eid] = diff
			}
		}
		// not enough
		if len(notenough) > 0 {
			err := &dots.Error{
				Code:    dots.ECONFLICT,
				Message: "not enough entries",
				Data:    map[string]interface{}{"notenough": notenough, "company_id": *upd.CompanyID},
			}
			return nil, err
		}
	}

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
		return nil, dots.Errorf(dots.ECONFLICT, "deed user conflict")
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

func (s *DeedService) DeleteDeed(ctx context.Context, id int, filter dots.DeedDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return deleteDeed(ctx, tx, id, filter)
	}

	if canerr := dots.CanDeleteOwn(ctx); canerr != nil {
		return 0, canerr
	}

	var n int
	// check search to own
	uid := dots.UserFromContext(ctx).ID

	err = deedBelongsToUser(ctx, tx, uid, id)
	if err != nil {
		return 0, err
	}

	n, err = deleteDeed(ctx, tx, id, filter)

	tx.Commit()

	return n, err
}

func createDeed(ctx context.Context, tx *Tx, d *dots.Deed) error {
	user := dots.UserFromContext(ctx)
	if user.ID == ksuid.Nil {
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

	if len(d.Distribute) == 0 {
		return nil
	}

	// manage distribute
	for eid, qty := range d.Distribute {
		d := dots.Drain{
			DeedID:    d.ID,
			EntryID:   eid,
			Quantity:  qty,
			IsDeleted: false,
		}

		err = createOrUpdateDrain(ctx, tx, d)
		if err != nil {
			// all or nothing
			return err
		}

	}

	return nil
}

func updateDeed(ctx context.Context, tx *Tx, id int, upd dots.DeedUpdate) (*dots.Deed, error) {
	dd, _, err := findDeed(ctx, tx, dots.DeedFilter{ID: &id, Limit: 1}, nil)
	if err != nil {
		return nil, fmt.Errorf("postgres.deed: cannot retrieve deed %w", err)
	}
	if len(dd) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "deed not found")
	}
	e := dd[0]

	set, args := []string{}, []interface{}{}
	if v := upd.Title; v != nil {
		e.Title = *v
		set, args = append(set, "title = ?"), append(args, *v)
	}
	if v := upd.Quantity; v != nil {
		e.Quantity = *v
		set, args = append(set, "quantity = ?"), append(args, *v)
	}
	if v := upd.Unit; v != nil {
		e.Unit = *v
		set, args = append(set, "unit = ?"), append(args, *v)
	}
	if v := upd.UnitPrice; v != nil {
		e.UnitPrice = *v
		set, args = append(set, "unitprice = ?"), append(args, *v)
	}
	if v := upd.CompanyID; v != nil {
		if e.CompanyID != *v {
			// start from fresh
			err = hardDeleteDrainsOfDeed(ctx, tx, e.ID)
			if err != nil {
				return nil, err
			}
		}
		e.CompanyID = *v
		set, args = append(set, "company_id = ?"), append(args, *v)
	}

	replaceQuestionMark(set, args)
	args = append(args, id)

	sqlstr := `
		update deed
		set ` + strings.Join(set, ", ") + `
		where	id = ` + fmt.Sprintf("$%d", len(args))

	_, err = tx.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.deed: cannot update %w", err)
	}

	if len(upd.Distribute) == 0 {
		return e, nil
	}

	// manage distribute need CompanyID
	if upd.Distribute != nil {
		if upd.CompanyID == nil {
			return nil, errors.New("company id is required")
		}
	}
	// delete all distribute
	err = deleteDrainsOfDeed(ctx, tx, e.ID)
	if err != nil {
		return nil, err
	}

	for eid, qty := range upd.Distribute {
		d := dots.Drain{
			DeedID:    e.ID,
			EntryID:   eid,
			Quantity:  qty,
			IsDeleted: false,
		}

		err = createOrUpdateDrain(ctx, tx, d)
		if err != nil {
			// all or nothing
			return nil, err
		}
	}

	e.Distribute = upd.Distribute

	err = hardDeleteDrainsOfDeedAlreadyDeleted(ctx, tx, e.ID)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func findDeed(ctx context.Context, tx *Tx, filter dots.DeedFilter, lockOwnID *ksuid.KSUID) (_ []*dots.Deed, n int, err error) {
	where, args := []string{}, []interface{}{}
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
	replaceQuestionMark(where, args)

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

func deleteDeed(ctx context.Context, tx *Tx, id int, filter dots.DeedDelete) (n int, err error) {
	where, args := []string{}, []interface{}{}
	where, args = append(where, "id = ?"), append(args, id)

	replaceQuestionMark(where, args)

	kind := "date_trunc('minute', now())::timestamptz"
	if filter.Resurect {
		kind = "null"
		where = append(where, "deleted_at is not null")
	} else {
		where = append(where, "deleted_at is null")
	}
	sqlstr := "update deed set deleted_at = " + kind + " where "
	sqlstr = sqlstr + strings.Join(where, " and ")

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.deed: cannot soft delete %w", err)
	}

	if filter.Undrain {
		err := undrainDrainsOfDeed(ctx, tx, id)
		if err != nil {
			return 0, fmt.Errorf("postgres.deed: cannot undrain %w", err)
		}
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func deedBelongsToUser(ctx context.Context, tx *Tx, u ksuid.KSUID, d int) error {
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

func deedGetUser(ctx context.Context, tx *Tx, d int) *ksuid.KSUID {
	sqlstr := `select c.tid
from company c
where c.id = (select d.company_id 
from deed d
where d.id = $1)
`
	var uid ksuid.KSUID
	err := tx.QueryRowContext(ctx, sqlstr, d).Scan(&uid)
	if err != nil {
		return nil
	}

	return &uid
}

// it builds a sql as next:
/*select json_object_agg(e.id, case when e.id = 54 then e.quantity - (select sum(case when d.is_deleted = true then -d.quantity else d.quantity end) from drain d where d.entry_id = e.id)... end) as enough from entry e where e.id = any(array[...]) and e.company_id = ...;*/
func entriesOfCompanyAreEnough(ctx context.Context, tx *Tx, eq map[int]float64, cid int) (map[int]float64, error) {

	var sqlb strings.Builder
	sqlb.WriteString("select json_object_agg( e.id, case ")
	ee := []interface{}{}
	placeholders := []string{}
	inx := 1

	for eid, quantity := range eq {
		sqlb.WriteString(fmt.Sprintf("when e.id = %d then (e.quantity - (select sum(case when d.is_deleted = true then -d.quantity else d.quantity end) from drain d where d.entry_id = e.id))  - %v ", eid, quantity))
		ee = append(ee, eid)

		placeholders = append(placeholders, fmt.Sprintf("$%d", inx))
		inx++
	}

	arr := strings.Join(placeholders, ", ")
	where := fmt.Sprintf(" end) as enough from entry e where e.id in (%s) and e.company_id = %d;", arr, cid)
	sqlb.WriteString(where)
	sqlstr := sqlb.String()

	// get byte representation of a json {int: flaot}
	var bb []byte
	err := tx.QueryRowContext(ctx, sqlstr, ee...).Scan(&bb)
	if err != nil {
		return nil, err
	}
	if bb == nil {
		return nil, sql.ErrNoRows
	}

	var check map[int]float64
	if err := json.Unmarshal(bb, &check); err != nil {
		return nil, err
	}

	return check, nil
}

func getEntryIDsFromDistribute(ee map[int]float64) []int {
	if len(ee) == 0 {
		return []int{}
	}

	ids := []int{}
	for id := range ee {
		ids = append(ids, id)
	}

	return ids
}

type entryRow struct {
	eid  int
	etid int
	qty  float64
}

func entriesOfEntryTypeForCompanyID(ctx context.Context, tx *Tx, etids []int, cid int) ([]entryRow, error) {
	sqlstr := `with s as (
  select e.id, e.entry_type_id, (e.quantity - coalesce((select sum(case when d.is_deleted = true then -d.quantity else d.quantity end)
from drain d
where d.entry_id = e.id), 0)
) quantity
from entry e
where e.entry_type_id = any(
		select et.id
from entry_type et
where et.id = any($1))
and
e.company_id = (
		select c.id
from company c
where c.id = $2 limit 1)
order by e.date_added  DESC
  ) select s.id, s.entry_type_id, s.quantity from s;`

	rows, err := tx.QueryContext(
		ctx,
		sqlstr,
		etids, cid,
	)
	if err == sql.ErrNoRows {
		return nil, dots.Errorf(dots.ENOTFOUND, "entries of entry not found")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		eid, etid int
		eqty      float64
		lines     []entryRow
	)
	for rows.Next() {
		err := rows.Scan(&eid, &etid, &eqty)
		if err != nil {
			return nil, err
		}
		line := entryRow{eid, etid, eqty}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
