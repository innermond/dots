package dots

import "context"

type Company struct {
	ID       int    `json:"id"`
	TID      int    `json:"tid"`
	Longname string `json:"longname"`
	TIN      string `json:"tin"`
	RN       string `json:"rn"`
}

func (c *Company) Validate() error {
	return nil
}

type CompanyFilter struct {
	ID       *int    `json:"id"`
	TID      *int    `json:"tid"`
	Longname *string `json:"longname"`
	TIN      *string `json:"tin"`
	RN       *string `json:"rn"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type CompanyService interface {
	// CreateCompany creates a company and
	// hydrate the input param to reflect changes, hence
	// the input param must be a pointer
	CreateCompany(context.Context, *Company) error
	UpdateCompany(context.Context, int, CompanyUpdate) (*Company, error)
	FindCompany(context.Context, CompanyFilter) ([]*Company, int, error)
}

type CompanyUpdate struct {
	TID      *int    `json:"tid"`
	Longname *string `json:"longname"`
	TIN      *string `json:"tin"`
	RN       *string `json:"rn"`
}

func (cu *CompanyUpdate) Valid() error {
	if cu.Longname == nil && cu.TIN == nil && cu.RN == nil {
		return Errorf(EINVALID, "at least name, tax identification number and  registration number are required")
	}

	return nil
}
