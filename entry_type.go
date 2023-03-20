package dots

import (
	"context"
)

type EntryType struct {
	ID          int     `json:"id"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	Unit        string  `json:"unit"`
	Tid         int     `json:"tid"`
}

func (et *EntryType) Validate() error {
	if len(et.Code) == 0 {
		return Errorf(EINVALID, "entry type code must not be empty")
	}

	if hasNonPrintable(et.Code) {
		return Errorf(EINVALID, "entry type code has non-printable characters")
	}

	return nil
}

type EntryTypeService interface {
	CreateEntryType(context.Context, *EntryType) error
	UpdateEntryType(context.Context, int, *EntryTypeUpdate) (*EntryType, error)
	FindEntryType(context.Context, *EntryTypeFilter) ([]*EntryType, int, error)
}

type EntryTypeFilter struct {
	ID          *int    `json:"id"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Unit        *string `json:"unit"`
	Tid         *int    `json:"tid"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type EntryTypeUpdate struct {
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Unit        *string `json:"unit"`
	Tid         *int    `json:"tid"`
}

func (etu *EntryTypeUpdate) Valid() error {
	if etu.Code == nil || etu.Unit == nil {
		return Errorf(EINVALID, "entry type code and unit are required")
	}
	if etu.Tid == nil {
		return Errorf(EINVALID, "entry type owner missing")
	}

	return nil
}
