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
	if err := c.Validate(); err != nil {
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

	if err := createCompany(ctx, tx, c); err != nil {
		return perr(err)
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

	if canerr := dots.CanReadOwn(ctx); canerr != nil {
		return nil, 0, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return nil, 0, err
	}

	return findCompany(ctx, tx, filter)
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id int, upd dots.CompanyUpdate) (*dots.Company, error) {
	if err := upd.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if canerr := dots.CanWriteOwn(ctx); canerr != nil {
		return nil, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return nil, err
	}

	c, err := updateCompany(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return c, nil
}

func (s *CompanyService) DeleteCompany(ctx context.Context, id int, filter dots.CompanyDelete) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if canerr := dots.CanDoAnything(ctx); canerr == nil {
		var n int
		var err error

		if filter.Hard {
			n, err = deleteCompanyPermanently(ctx, tx, id)
		} else {
			n, err = deleteCompany(ctx, tx, id, filter.Resurect)
		}
		if err != nil {
			return n, err
		}

		tx.Commit()

		return n, err
	}

	if canerr := dots.CanDeleteOwn(ctx); canerr != nil {
		return 0, canerr
	}

	if err := tx.setUserIDPerConnection(ctx); err != nil {
		return 0, err
	}

	var n int

	if filter.Hard {
		n, err = deleteCompanyPermanently(ctx, tx, id)
	} else {
		n, err = deleteCompany(ctx, tx, id, filter.Resurect)
	}
	if err != nil {
		return n, err
	}

	tx.Commit()

	return n, err
}

func findCompany(ctx context.Context, tx *Tx, filter dots.CompanyFilter) (_ []*dots.Company, n int, err error) {
	where, args := []string{}, []interface{}{}
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

	// TODO deal with isDeleted
	/*	v := filter.IsDeleted
		if v != nil && *v {
			where = append(where, "deleted_at is not null")
		} else if v != nil && !*v {
			where = append(where, "deleted_at is null")
		} else if v == nil {
			where = append(where, "deleted_at is null")
		}
	*/
	wherestr := ""
	if len(where) > 0 {
		replaceQuestionMark(where, args)
		wherestr = "where " + strings.Join(where, " and ")
	}
	sqlstr := `
		select id, longname, tin, rn, count(*) over() from company
		` + wherestr + ` ` + formatLimitOffset(filter.Limit, filter.Offset)
	rows, err := tx.QueryContext(
		ctx,
		sqlstr,
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
		err := rows.Scan(&e.ID, &e.Longname, &e.TIN, &e.RN, &n)
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
	sqlstr, args := `
insert into company
(longname, tin, rn)
values
($1, $2, $3) returning id
	`, []interface{}{c.Longname, c.TIN, c.RN}

	if err := tx.QueryRowContext(
		ctx,
		sqlstr,
		args...,
	).Scan(&c.ID); err != nil {
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

func deleteCompany(ctx context.Context, tx *Tx, id int, resurect bool) (n int, err error) {
	where := []string{"core.company.id = $1"}

	kind := "date_trunc('minute', now())::timestamptz"
	if resurect {
		kind = "null"
		where = append(where, "core.company.deleted_at is not null")
	} else {
		where = append(where, "core.company.deleted_at is null")
	}

	wherestr := "where " + strings.Join(where, " and ")

	bareCompany := `
	and not exists(
		select c.id from core.company c join core.entry e on(c.id = e.company_id)
		where e.company_id = $1 limit 1) 
	and not exists(
		select c.id from core.company c join core.deed d on(c.id = d.company_id)
		where d.company_id = $1 limit 1)`

	sqlstr := `update core.company set deleted_at = %s ` + wherestr + bareCompany
	sqlstr = fmt.Sprintf(sqlstr, kind)

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		id,
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

func deleteCompanyPermanently(ctx context.Context, tx *Tx, id int) (n int, err error) {
	where := []string{"core.company.id = $1"}
	wherestr := "where " + strings.Join(where, " and ")

	bareCompany := `
	and not exists(
		select c.id from core.company c join core.entry e on(c.id = e.company_id)
		where e.company_id = $1 limit 1) 
	and not exists(
		select c.id from core.company c join core.deed d on(c.id = d.company_id)
		where d.company_id = $1 limit 1)`

	sqlstr := `delete from core.company ` + wherestr + bareCompany

	result, err := tx.ExecContext(
		ctx,
		sqlstr,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres.company: cannot hard delete %w", err)
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
from core.company c
where c.id = $1 and c.tid = $2
);
`
	var exists bool
	err := tx.QueryRowContext(ctx, sqlstr, companyID, u).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return dots.Errorf(dots.ENOTFOUND, "company not found")
	}

	return nil
}
