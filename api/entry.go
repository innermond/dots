package dots

import (
	"context"
	"time"
)

type Entry struct {
	ID          int       `json:"id"`
	EntryTypeID int       `json:"entry_type_id"`
	DateAdded   time.Time `json:"date_added"`
	Quantity    float64   `json:"quantity"`
	CompanyID   int       `json:"company_id"`
}

func (e *Entry) Validate() error {
	if e.EntryTypeID <= 0 || int(e.Quantity) <= 0 || e.CompanyID <= 0 {
		return Errorf(EINVALID, "zero or bellow is not accepted")
	}
	return nil
}

type EntryService interface {
	CreateEntry(context.Context, *Entry) error
	UpdateEntry(context.Context, int, EntryUpdate) (*Entry, error)
	FindEntry(context.Context, EntryFilter) ([]*Entry, int, error)
	DeleteEntry(context.Context, int, EntryDelete) (int, error)
}

type EntryFilter struct {
	ID          *int       `json:"id"`
	EntryTypeID *int       `json:"entry_type_id"`
	DateAdded   *time.Time `json:"date_added"`
	Quantity    *float64   `json:"quantity"`
	CompanyID   *int       `json:"company_id"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	IsDeleted     bool         `json:"is_deleted"`
	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type EntryDelete struct {
	Resurect bool
}

type EntryUpdate struct {
	EntryTypeID *int       `json:"entry_type_id"`
	DateAdded   *time.Time `json:"date_added"`
	Quantity    *float64   `json:"quantity"`
	CompanyID   *int       `json:"company_id"`
}

func (eu *EntryUpdate) Valid() error {
	if eu.EntryTypeID == nil || eu.Quantity == nil || eu.CompanyID == nil {
		return Errorf(EINVALID, "entry type, quantity and company are required")
	}

	return nil
}
