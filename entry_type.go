package dots

import (
	"context"

	"github.com/segmentio/ksuid"
)

type EntryType struct {
	ID          int         `json:"id"`
	Code        string      `json:"code"`
	Description *string     `json:"description"`
	Unit        string      `json:"unit"`
	TID         ksuid.KSUID `json:"tid"`
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
	UpdateEntryType(context.Context, int, EntryTypeUpdate) (*EntryType, error)
	FindEntryType(context.Context, EntryTypeFilter) ([]*EntryType, int, error)
	DeleteEntryType(context.Context, EntryTypeDelete) (int, error)
}

type EntryTypeFilter struct {
	ID          *int         `json:"id"`
	Code        *string      `json:"code"`
	Description *string      `json:"description"`
	Unit        *string      `json:"unit"`
	TID         *ksuid.KSUID `json:"tid"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type EntryTypeDelete struct {
	EntryTypeFilter

	Resurect bool
}

type EntryTypeUpdate struct {
	Code        *string      `json:"code"`
	Description *string      `json:"description"`
	Unit        *string      `json:"unit"`
	TID         *ksuid.KSUID `json:"tid"`
}

func (etu *EntryTypeUpdate) Valid() error {
	if etu.Code == nil && etu.Unit == nil && etu.Description == nil {
		return Errorf(EINVALID, "entry type code or unit or description are required")
	}
	if etu.TID == nil {
		return Errorf(EINVALID, "entry type owner missing")
	}

	return nil
}
