package dots

import "context"

type EntryType struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	Tid         int    `json:"tid"`
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
}
