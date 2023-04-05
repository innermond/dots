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

	if err := createDeed(ctx, tx, d); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (s *DeedService) FindDeed(ctx context.Context, filter *dots.DeedFilter) ([]*dots.Deed, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	return findDeed(ctx, tx, filter)
}

func (s *DeedService) UpdateDeed(ctx context.Context, id int, upd dots.DeedUpdate) (*dots.Deed, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	d, err := updateDeed(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return d, nil
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

	return nil
}

func updateDeed(ctx context.Context, tx *Tx, id int, updata dots.DeedUpdate) (*dots.Deed, error) {
	dd, _, err := findDeed(ctx, tx, &dots.DeedFilter{ID: &id, Limit: 1})
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

	return e, nil
}

func findDeed(ctx context.Context, tx *Tx, filter *dots.DeedFilter) (_ []*dots.Deed, n int, err error) {
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
	for inx, v := range where {
		if !strings.Contains(v, "?") {
			continue
		}
		v = strings.Replace(v, "?", fmt.Sprintf("$%d", inx), 1)
		where[inx] = v
	}

	rows, err := tx.QueryContext(ctx, `
		select id, title, unit, unitprice, quantity, company_id, count(*) over() from deed
		where `+strings.Join(where, " and ")+` `+formatLimitOffset(filter.Limit, filter.Offset),
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
