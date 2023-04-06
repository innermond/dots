package dots

import "context"

type Drain struct {
	DeedID   int     `json:"deed_id"`
	EntryID  int     `json:"entry_id"`
	Quantity float64 `json:"quantity"`
}

func (d *Drain) Validate() error {
	return nil
}

type DrainService interface {
	CreateDrain(context.Context, Drain) error
	//UpdateDrain(context.Context, int, *Drain) (*Drain, error)
	//FindDrain(context.Context, *DrainFilter) ([]*Drain, int, error)
}

type DrainFilter struct {
	DeedID   *int     `json:"deed_id"`
	EntryID  *int     `json:"entry_id"`
	Quantity *float64 `json:"quantity"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type DrainUpdate struct {
	DeedID   *int     `json:"deed_id"`
	EntryID  *int     `json:"entry_id"`
	Quantity *float64 `json:"quantity"`
}

func (d *DrainUpdate) Valid() error {
	if d.DeedID == nil && d.EntryID == nil && d.Quantity == nil {
		return Errorf(EINVALID, "all drain data is nil")
	}

	return nil
}
