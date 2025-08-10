package entities

import (
	"time"

	"github.com/google/uuid"
)

type AddressType string

const (
	AddressTypeHome     AddressType = "home"
	AddressTypeWork     AddressType = "work"
	AddressTypeBilling  AddressType = "billing"
	AddressTypeShipping AddressType = "shipping"
	AddressTypeOther    AddressType = "other"
)

type AddressableType string

const (
	AddressableTypeUser         AddressableType = "user"
	AddressableTypeOrganization AddressableType = "organization"
)

type Address struct {
	ID              uuid.UUID       `json:"id"`
	AddressableID   uuid.UUID       `json:"addressable_id"`
	AddressableType AddressableType `json:"addressable_type"`
	AddressLine1    string          `json:"address_line1"`
	AddressLine2    *string         `json:"address_line2,omitempty"`
	City            string          `json:"city"`
	State           string          `json:"state"`
	PostalCode      string          `json:"postal_code"`
	Country         string          `json:"country"`
	Latitude        *float64        `json:"latitude,omitempty"`
	Longitude       *float64        `json:"longitude,omitempty"`
	IsPrimary       bool            `json:"is_primary"`
	AddressType     AddressType     `json:"address_type"`
	CreatedBy       uuid.UUID       `json:"created_by"`
	UpdatedBy       uuid.UUID       `json:"updated_by"`
	DeletedBy       *uuid.UUID      `json:"deleted_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty"`
}

func NewAddress(
	addressableID uuid.UUID,
	addressableType AddressableType,
	addressLine1, city, state, postalCode, country string,
	addressLine2 *string,
	latitude, longitude *float64,
	isPrimary bool,
	addressType AddressType,
) *Address {
	return &Address{
		ID:              uuid.New(),
		AddressableID:   addressableID,
		AddressableType: addressableType,
		AddressLine1:    addressLine1,
		AddressLine2:    addressLine2,
		City:            city,
		State:           state,
		PostalCode:      postalCode,
		Country:         country,
		Latitude:        latitude,
		Longitude:       longitude,
		IsPrimary:       isPrimary,
		AddressType:     addressType,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (a *Address) UpdateAddress(
	addressLine1, city, state, postalCode, country string,
	addressLine2 *string,
	latitude, longitude *float64,
	isPrimary bool,
	addressType AddressType,
) {
	a.AddressLine1 = addressLine1
	a.AddressLine2 = addressLine2
	a.City = city
	a.State = state
	a.PostalCode = postalCode
	a.Country = country
	a.Latitude = latitude
	a.Longitude = longitude
	a.IsPrimary = isPrimary
	a.AddressType = addressType
	a.UpdatedAt = time.Now()
}

func (a *Address) SetPrimary(isPrimary bool) {
	a.IsPrimary = isPrimary
	a.UpdatedAt = time.Now()
}

func (a *Address) SetAddressType(addressType AddressType) {
	a.AddressType = addressType
	a.UpdatedAt = time.Now()
}

func (a *Address) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}

func (a *Address) IsDeleted() bool {
	return a.DeletedAt != nil
}

func (a *Address) IsPrimaryAddress() bool {
	return a.IsPrimary
}

func (a *Address) GetFullAddress() string {
	address := a.AddressLine1
	if a.AddressLine2 != nil && *a.AddressLine2 != "" {
		address += ", " + *a.AddressLine2
	}
	address += ", " + a.City + ", " + a.State + " " + a.PostalCode + ", " + a.Country
	return address
}

func (a *Address) HasCoordinates() bool {
	return a.Latitude != nil && a.Longitude != nil
}
