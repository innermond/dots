package postgres

import (
	"errors"
)

var (
	ErrDistributeEmpty       = errors.New("empty distribute")
	ErrDistributeNotEnough   = errors.New("not enough quantities")
	ErrDistributeCalculation = errors.New("not found a solution")
)

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
