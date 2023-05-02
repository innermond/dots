package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/innermond/dots"
	"github.com/segmentio/ksuid"
)

type CompanyService struct {
	db *DB
}

func NewCompanyService(db *DB) *CompanyService {
	return &CompanyService{db: db}
}

func (s *CompanyService) CreateCompany(ctx context.Context, c *dots.Company) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return createCompany(ctx, tx, c)
	}

	if canerr := dots.CanCreateOwn(ctx); canerr != nil {
		return canerr
	}

	if err := createCompany(ctx, tx, c); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (s *CompanyService) FindCompany(ctx context.Context, filter dots.CompanyFilter) ([]*dots.Company, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return findCompany(ctx, tx, filter)
	}

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	uid := dots.UserFromContext(ctx).ID
	// trying to get companies for a different TID
	if filter.TID != nil && *filter.TID != uid {
		// will get empty results and not error
		return make([]*dots.Company, 0), 0, nil
	}
	// lock search to own
	filter.TID = &uid

	return findCompany(ctx, tx, filter)
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id int, upd dots.CompanyUpdate) (*dots.Company, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return updateCompany(ctx, tx, id, upd)
	}

	uid := dots.UserFromContext(ctx).ID
	err = companyBelongsToUser(ctx, tx, uid, id)
	if err != nil {
		return nil, err
	}

	if upd.TID != nil && uid != *upd.TID {
		return nil, dots.Errorf(dots.EINVALID, "update company: unexpected user")
	}

	if canerr := dots.CanWriteOwn(ctx, *upd.TID); canerr != nil {
		return nil, canerr
	}

	c, err := updateCompany(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return c, nil
}

func (s *CompanyService) DeleteCompany(ctx context.Context, filter dots.CompanyDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		return deleteCompany(ctx, tx, filter)
	}

	if canerr := dots.CanDeleteOwn(ctx); canerr != nil {
		return 0, canerr
	}

	var n int
	// lock delete to own
	uid := dots.UserFromContext(ctx).ID
	filter.TID = &uid

	if filter.ID != nil {
		err = companyBelongsToUser(ctx, tx, uid, *filter.ID)
		if err != nil {
			return 0, err
		}
		n, err = deleteCompany(ctx, tx, filter)
	} else {
		n, err = deleteCompany(ctx, tx, filter)
	}

	tx.Commit()

	return n, err
}

func findCompany(ctx context.Context, tx *Tx, filter dots.CompanyFilter) (_ []*dots.Company, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Longname; v != nil {
		where, args = append(where, "longname = ?"), append(args, *v)
	}
	if v := filter.TIN; v != nil {
		where, args = append(where, "tin = ?"), append(args, *v)
	}
	if v := filter.RN; v != nil {
		where, args = append(where, "rn = ?"), append(args, *v)
	}
	if v := filter.TID; v != nil {
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
		select id, longname, tin, rn, tid, count(*) over() from company
		where `+strings.Join(where, " and ")+` `+formatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err == sql.ErrNoRows {
		return nil, 0, dots.Errorf(dots.ENOTFOUND, "company not found")
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	companies := []*dots.Company{}
	for rows.Next() {
		var e dots.Company
		err := rows.Scan(&e.ID, &e.Longname, &e.TIN, &e.RN, &e.TID, &n)
		if err != nil {
			return nil, 0, err
		}
		companies = append(companies, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return companies, n, nil
}

func createCompany(ctx context.Context, tx *Tx, c *dots.Company) error {
	user := dots.UserFromContext(ctx)
	if user.ID == ksuid.Nil {
		return dots.Errorf(dots.EUNAUTHORIZED, "unauthorized user")
	}

	if err := c.Validate(); err != nil {
		return err
	}

	err := tx.QueryRowContext(
		ctx,
		`
insert into company
(longname, tin, rn, tid)
values
($1, $2, $3, $4) returning id
		`,
		c.Longname, c.TIN, c.RN, user.ID,
	).Scan(&c.ID)
	if err != nil {
		return err
	}

	return nil
}

func updateCompany(ctx context.Context, tx *Tx, id int, updata dots.CompanyUpdate) (*dots.Company, error) {
	cc, _, err := findCompany(ctx, tx, dots.CompanyFilter{ID: &id, Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("postgres.company: cannot retrieve company type %w", err)
	}
	if len(cc) == 0 {
		return nil, dots.Errorf(dots.ENOTFOUND, "company not found")
	}
	ct := cc[0]

	set, args := []string{}, []interface{}{}
	if v := updata.Longname; v != nil {
		ct.Longname = *v
		set, args = append(set, "longname = ?"), append(args, *v)
	}
	if v := updata.TIN; v != nil {
		ct.TIN = *v
		set, args = append(set, "tin = ?"), append(args, *v)
	}
	if v := updata.RN; v != nil {
		ct.RN = *v
		set, args = append(set, "rn = ?"), append(args, *v)
	}

	for inx, v := range set {
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx+1), 1)
		set[inx] = v
	}
	args = append(args, id)

	sqlstr := `
		update company
		set ` + strings.Join(set, ", ") + `
		where	id = ` + fmt.Sprintf("$%d", len(args))

	_, err = tx.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.company: cannot update %w", err)
	}

	return ct, nil
}

func deleteCompany(ctx context.Context, tx *Tx, filter dots.CompanyDelete) (n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "c.id = ?"), append(args, *v)
	}
	if v := filter.Longname; v != nil {
		where, args = append(where, "longname = ?"), append(args, *v)
	}
	if v := filter.TIN; v != nil {
		where, args = append(where, "tin = ?"), append(args, *v)
	}
	if v := filter.RN; v != nil {
		where, args = append(where, "rn = ?"), append(args, *v)
	}
	if v := filter.TID; v != nil {
		where, args = append(where, "tid = ?"), append(args, *v)
	}
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}

	kind := "date_trunc('minute', now())::timestamptz"
	if filter.Resurect {
		kind = "null"
		where = append(where, "c.deleted_at is not null")
	} else {
		where = append(where, "c.deleted_at is null")
	}

	whereEntries, whereDeeds := make([]string, len(where)), make([]string, len(where))
	copy(whereEntries, where)
	copy(whereDeeds, where)
	whereEntries = append(whereEntries, "e.company_id is null")
	whereDeeds = append(whereDeeds, "d.company_id is null")

	sqlstr := `
		update company set deleted_at = %s where id = any(
		select c.id from company c left join entry e on(c.id = e.company_id)
		where %s) or id = any(
		select c.id from company c left join deed d on(c.id = d.company_id)
		where %s)`
	conditionEntries := strings.Join(whereEntries, " and ") + ` ` + formatLimitOffset(filter.Limit, filter.Offset)
	conditionDeeds := strings.Join(whereDeeds, " and ") + ` ` + formatLimitOffset(filter.Limit, filter.Offset)
	sqlstr = fmt.Sprintf(sqlstr, kind, conditionEntries, conditionDeeds)
	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		args...,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.company: cannot soft delete %w", err)
	}

	n64, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func companyBelongsToUser(ctx context.Context, tx *Tx, u ksuid.KSUID, companyID int) error {
	sqlstr := `select exists(
select id
from company c
where c.id = $1 and c.tid = $2
);
`
	var exists bool
	err := tx.QueryRowContext(ctx, sqlstr, companyID, u).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return dots.Errorf(dots.EUNAUTHORIZED, "foreign entry")
	}

	return nil
}
