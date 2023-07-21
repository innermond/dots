package dots

import (
	"context"
	"regexp"
	"strings"

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

	// trim white space
	if c.Longname != "" {
		c.Longname = strings.Trim(c.Longname, " ")
	}
	if c.TIN != "" {
		c.TIN = strings.Trim(c.TIN, " ")
	}
	if c.RN != "" {
		c.RN = strings.Trim(c.RN, " ")
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

	IsDeleted *bool `json:"is_deleted"`

	DeletedAtFrom *PartialTime `json:"deleted_at_from,omitempty"`
	DeletedAtTo   *PartialTime `json:"deleted_at_to,omitempty"`
}

type CompanyDelete struct {
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
	DeleteCompany(context.Context, int, CompanyDelete) (int, error)
}

type CompanyUpdate struct {
	TID      *ksuid.KSUID `json:"tid"`
	Longname *string      `json:"longname"`
	TIN      *string      `json:"tin"`
	RN       *string      `json:"rn"`
}

func (cu *CompanyUpdate) Validate() error {
	// required
	if cu.Longname == nil && cu.TIN == nil && cu.RN == nil {
		return Errorf(EINVALID, "al least one of name, tax identification number or registration number are required")
	}

	// trim white space
	if cu.Longname != nil {
		if len(*cu.Longname) == 0 {
			return Errorf(EINVALID, "name need content")
		}
		*cu.Longname = strings.Trim(*cu.Longname, " ")
	}
	if cu.TIN != nil {
		if len(*cu.TIN) == 0 {
			return Errorf(EINVALID, "tax identification  number need content")
		}
		*cu.TIN = strings.Trim(*cu.TIN, " ")
	}
	if cu.RN != nil {
		if len(*cu.RN) == 0 {
			return Errorf(EINVALID, "registration number need content")
		}
		*cu.RN = strings.Trim(*cu.RN, " ")
	}

	return nil
}
