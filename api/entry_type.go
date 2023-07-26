package dots

import (
	"context"
)

type EntryType struct {
	ID          int     `json:"id"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	Unit        string  `json:"unit"`
}

func (et *EntryType) Validate() error {
	if len(et.Code) == 0 {
		return Errorf(EINVALID, "entry type code must not be empty")
	}

	suspects := []*string{&et.Code, et.Description, &et.Unit}
	printable(suspects)

	return nil
}

type EntryTypeService interface {
	CreateEntryType(context.Context, *EntryType) error
	UpdateEntryType(context.Context, int, EntryTypeUpdate) (*EntryType, error)
	FindEntryType(context.Context, EntryTypeFilter) ([]*EntryType, int, error)
	DeleteEntryType(context.Context, int, EntryTypeDelete) (int, error)
}

type EntryTypeFilter struct {
	ID          *int    `json:"id"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Unit        *string `json:"unit"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	IsDeleted *bool `json:"is_deleted"`

	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type EntryTypeDelete struct {
	Resurect bool
}

type EntryTypeUpdate struct {
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Unit        *string `json:"unit"`
}

func (etu *EntryTypeUpdate) Validate() error {
	if etu.Code == nil && etu.Unit == nil && etu.Description == nil {
		return Errorf(EINVALID, "entry type code or unit or description are required")
	}

	return nil
}
