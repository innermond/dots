package dots

import (
	"context"
	"regexp"

	"github.com/segmentio/ksuid"
)

type Company struct {
	ID       int         `json:"id"`
	TID      ksuid.KSUID `json:"tid"`
	Longname string      `json:"longname"`
	TIN      string      `json:"tin"`
	RN       string      `json:"rn"`
}

func (c *Company) Validate() error {
	if c.Longname == "" || c.TIN == "" || c.RN == "" {
		return Errorf(EINVALID, "all name, tax identification number and  registration number are required")
	}

  suspects := []string{c.Longname, c.TIN, c.RN}
  // all utf-8 except control charatcters
  pattern := "^[[:^cntrl:]]+$"
  re := regexp.MustCompile(pattern)
  for _, suspect := range suspects {
    match := re.MatchString(suspect)
    if !match {
      return Errorf(EINVALID, "input is not a text line")
    }
  }
  // only white spaces
  pattern = "^\\s+$"
  re = regexp.MustCompile(pattern)
  for _, suspect := range suspects {
    match := re.MatchString(suspect)
    if match {
      return Errorf(EINVALID, "emptyness as input")
    }
  }

	return nil
}

type CompanyFilter struct {
	ID       *int         `json:"id"`
	TID      *ksuid.KSUID `json:"tid"`
	Longname *string      `json:"longname"`
	TIN      *string      `json:"tin"`
	RN       *string      `json:"rn"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

  IsDeleted *bool

	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type CompanyDelete struct {
	CompanyFilter

  // delete will be hard using "delete" sql kwyword
  Hard bool
  // update deletion field
	Resurect bool
}

type CompanyService interface {
	// CreateCompany creates a company and
	// hydrate the input param to reflect changes, hence
	// the input param must be a pointer
	CreateCompany(context.Context, *Company) error
	UpdateCompany(context.Context, int, CompanyUpdate) (*Company, error)
	FindCompany(context.Context, CompanyFilter) ([]*Company, int, error)
	DeleteCompany(context.Context, CompanyDelete) (int, error)
}

type CompanyUpdate struct {
	TID      *ksuid.KSUID `json:"tid"`
	Longname *string      `json:"longname"`
	TIN      *string      `json:"tin"`
	RN       *string      `json:"rn"`
}

func (cu *CompanyUpdate) Valid() error {
	if cu.Longname == nil && cu.TIN == nil && cu.RN == nil {
		return Errorf(EINVALID, "at least name, tax identification number and  registration number are required")
	}

	return nil
}
