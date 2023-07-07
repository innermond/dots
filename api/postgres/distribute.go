package postgres

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrDistributeEmpty       = errors.New("empty distribute")
	ErrDistributeNotEnough   = errors.New("not enough quantities")
	ErrDistributeCalculation = errors.New("not found a solution")
)

// TODO this is no longer used as will force calling code to use a map that
// will alter the order given from database
func DistributeFrom(dd map[int]map[int]float64, etd map[int]float64) (map[int]float64, error) {

	if len(dd) == 0 {
		return nil, ErrDistributeEmpty
	}

	distribute := map[int]map[int]float64{}
	for etid, requiredOty := range etd {
		idqty, found := dd[etid]
		if !found {
			continue
		}
		entryOty := map[int]float64{}
		for id, qty := range idqty {
			if requiredOty == 0.0 {
				continue
			}
			// enough case
			if requiredOty <= qty {
				entryOty[id] = requiredOty
				// it is completly consumed
				// this consuming state will be checked later in code
				requiredOty = 0
				break
			}
			// not enough need more entries to consume
			requiredOty -= qty
			entryOty[id] = qty
		}
		distribute[etid] = entryOty
		// hasn't been consumed
		if requiredOty > 0 {
			return nil, ErrDistributeNotEnough
		}
	}

	// found no solution
	if len(distribute) == 0 {
		return nil, ErrDistributeEmpty
	}

	calculated := map[int]float64{}
	for _, eidqty := range distribute {
		for eid, qty := range eidqty {
			calculated[eid] = qty
		}
	}

	if len(calculated) == 0 {
		return nil, ErrDistributeCalculation
	}

	return calculated, nil
}

func suggestDistributeOverEntryType(ctx context.Context, tx *Tx, etid int, num float64) (map[int]float64, error) {
	sqlstr := `
with entry as (
 select e.date_added, e.id, e.entry_type_id, (e.quantity - coalesce((select sum(case when d.is_deleted = true then -d.quantity else d.quantity end)
from drain d
where d.entry_id = e.id), 0)
) quantity
from entry e
where e.entry_type_id = $1
order by e.date_added desc  
), cumulative_sum as (
  select id, quantity, date_added, SUM(quantity) over (partition by entry_type_id order by date_added desc) as running_sum
from entry 
)
select id, quantity, case
    when running_sum <= $2 then quantity
else quantity - (running_sum - $2)
  end as subtracted_quantity,
  running_sum
from cumulative_sum
where quantity - (running_sum - $2) >= 0
;
	`

	var m map[int]float64
	err := tx.QueryRowContext(ctx, sqlstr, etid, num).Scan(&m)
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, sql.ErrNoRows
	}

	return m, nil
}
