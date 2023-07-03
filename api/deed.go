package dots

import (
	"context"

	"github.com/shopspring/decimal"
)

type Deed struct {
	ID        int             `json:"id"`
	CompanyID int             `json:"company_id"`
	Title     string          `json:"title"`
	Quantity  float64         `json:"quantity"`
	Unit      string          `json:"unit"`
	UnitPrice decimal.Decimal `json:"unitprice"`

	Distribute map[int]float64 `json:"distribute"`

	EntryTypeID        *int             `json:"entry_type_id,omitempty"`
	DistributeStrategy *DistributeDrain `json:"distribute_strategy"`
}

type DistributeDrain int

const (
	DistributeFromOldest DistributeDrain = iota
	DistributeFromNewest
	DistributeAsEqual
)

func (d *Deed) Validate() error {
	return nil
}

type DeedService interface {
	CreateDeed(context.Context, *Deed) error
	UpdateDeed(context.Context, int, DeedUpdate) (*Deed, error)
	FindDeed(context.Context, DeedFilter) ([]*Deed, int, error)
	DeleteDeed(context.Context, int, DeedDelete) (int, error)
}

type DeedFilter struct {
	ID        *int             `json:"id"`
	CompanyID *int             `json:"company_id"`
	Title     *string          `json:"title"`
	Quantity  *float64         `json:"quantity"`
	Unit      *string          `json:"unit"`
	UnitPrice *decimal.Decimal `json:"unitprice"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type DeedDelete struct {
	Resurect bool
}

type DeedUpdate struct {
	CompanyID *int             `json:"company_id"`
	Title     *string          `json:"title"`
	Quantity  *float64         `json:"quantity"`
	Unit      *string          `json:"unit"`
	UnitPrice *decimal.Decimal `json:"unitprice"`

	Distribute map[int]float64 `json:"distribute"`

	EntryID         *int     `json:"entry_id"`
	DrainedQuantity *float64 `json:"drained_quantity"`
}

func (du *DeedUpdate) Valid() error {
	if du.Title == nil && du.Quantity == nil && du.Unit == nil {
		return Errorf(EINVALID, "at least title, quantity and unit are required")
	}

	return nil
}
