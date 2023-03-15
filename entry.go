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
	return nil
}

type EntryService interface {
	CreateEntry(context.Context, *Entry) error
}

type EntryFilter struct {
	ID          *int       `json:"id"`
	EntryTypeID *int       `json:"entry_type_id"`
	DateAdded   *time.Time `json:"date_added"`
	Quantity    *float64   `json:"quantity"`
	CompanyID   *int       `json:"company_id"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
