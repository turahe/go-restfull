package requests

import (
	"github.com/go-playground/validator/v10"
)

// CreateAddressRequest represents the request for creating an address
type CreateAddressRequest struct {
	AddressableID   string   `json:"addressable_id" validate:"required,uuid4"`
	AddressableType string   `json:"addressable_type" validate:"required,oneof=user organization"`
	AddressLine1    string   `json:"address_line1" validate:"required,min=1,max=255"`
	AddressLine2    *string  `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City            string   `json:"city" validate:"required,min=1,max=255"`
	State           string   `json:"state" validate:"required,min=1,max=255"`
	PostalCode      string   `json:"postal_code" validate:"required,min=1,max=20"`
	Country         string   `json:"country" validate:"required,min=1,max=255"`
	Latitude        *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude       *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	IsPrimary       bool     `json:"is_primary"`
	AddressType     string   `json:"address_type" validate:"required,oneof=home work billing shipping other"`
}

// Validate validates the CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// UpdateAddressRequest represents the request for updating an address
type UpdateAddressRequest struct {
	AddressLine1 string   `json:"address_line1,omitempty" validate:"omitempty,min=1,max=255"`
	AddressLine2 *string  `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City         string   `json:"city,omitempty" validate:"omitempty,min=1,max=255"`
	State        string   `json:"state,omitempty" validate:"omitempty,min=1,max=255"`
	PostalCode   string   `json:"postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	Country      string   `json:"country,omitempty" validate:"omitempty,min=1,max=255"`
	Latitude     *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude    *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	IsPrimary    *bool    `json:"is_primary,omitempty"`
	AddressType  string   `json:"address_type,omitempty" validate:"omitempty,oneof=home work billing shipping other"`
}

// Validate validates the UpdateAddressRequest
func (r *UpdateAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// SetPrimaryAddressRequest represents the request for setting an address as primary
type SetPrimaryAddressRequest struct {
	AddressableID   string `json:"addressable_id" validate:"required,uuid4"`
	AddressableType string `json:"addressable_type" validate:"required,oneof=user organization"`
}

// Validate validates the SetPrimaryAddressRequest
func (r *SetPrimaryAddressRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// SetAddressTypeRequest represents the request for setting an address type
type SetAddressTypeRequest struct {
	AddressType string `json:"address_type" validate:"required,oneof=home work billing shipping other"`
}

// Validate validates the SetAddressTypeRequest
func (r *SetAddressTypeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// SearchAddressesRequest represents the request for searching addresses
type SearchAddressesRequest struct {
	City       string `json:"city,omitempty" validate:"omitempty,min=1,max=255"`
	State      string `json:"state,omitempty" validate:"omitempty,min=1,max=255"`
	Country    string `json:"country,omitempty" validate:"omitempty,min=1,max=255"`
	PostalCode string `json:"postal_code,omitempty" validate:"omitempty,min=1,max=20"`
	Limit      int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset     int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// Validate validates the SearchAddressesRequest
func (r *SearchAddressesRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Set default values
	if r.Limit == 0 {
		r.Limit = 10
	}

	return nil
}
