package controllers

import (
	"net/http"
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AddressController struct {
	addressService ports.AddressService
}

func NewAddressController(addressService ports.AddressService) *AddressController {
	return &AddressController{
		addressService: addressService,
	}
}

// CreateAddress godoc
//	@Summary		Create a new address
//	@Description	Create a new address for a user or organization
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateAddressRequest						true	"Address creation request"
//	@Success		201		{object}	responses.SuccessResponse{data=entities.Address}	"Address created successfully"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid input data"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses [post]
//	@Security		BearerAuth
func (c *AddressController) CreateAddress(ctx *fiber.Ctx) error {
	var req requests.CreateAddressRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	addressableID, err := uuid.Parse(req.AddressableID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable ID format",
		})
	}

	addressableType := entities.AddressableType(req.AddressableType)
	addressType := entities.AddressType(req.AddressType)

	address, err := c.addressService.CreateAddress(
		ctx.Context(),
		addressableID,
		addressableType,
		req.AddressLine1,
		req.City,
		req.State,
		req.PostalCode,
		req.Country,
		req.AddressLine2,
		req.Latitude,
		req.Longitude,
		req.IsPrimary,
		addressType,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   address,
	})
}

// GetAddressByID godoc
//	@Summary		Get address by ID
//	@Description	Retrieve an address by its unique identifier
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Address ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse{data=entities.Address}	"Address found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid address ID format"
//	@Failure		404	{object}	responses.ErrorResponse								"Address not found"
//	@Router			/api/v1/addresses/{id} [get]
//	@Security		BearerAuth
func (c *AddressController) GetAddressByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address ID format",
		})
	}

	address, err := c.addressService.GetAddressByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Address not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   address,
	})
}

// UpdateAddress godoc
//	@Summary		Update an address
//	@Description	Update an existing address by ID
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string												true	"Address ID"	format(uuid)
//	@Param			request	body		requests.UpdateAddressRequest						true	"Address update request"
//	@Success		200		{object}	responses.SuccessResponse{data=entities.Address}	"Address updated successfully"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid input data"
//	@Failure		404		{object}	responses.ErrorResponse								"Address not found"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses/{id} [put]
//	@Security		BearerAuth
func (c *AddressController) UpdateAddress(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address ID format",
		})
	}

	var req requests.UpdateAddressRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Get current address to preserve unchanged fields
	currentAddress, err := c.addressService.GetAddressByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Address not found",
		})
	}

	// Update only provided fields
	addressLine1 := req.AddressLine1
	if addressLine1 == "" {
		addressLine1 = currentAddress.AddressLine1
	}

	city := req.City
	if city == "" {
		city = currentAddress.City
	}

	state := req.State
	if state == "" {
		state = currentAddress.State
	}

	postalCode := req.PostalCode
	if postalCode == "" {
		postalCode = currentAddress.PostalCode
	}

	country := req.Country
	if country == "" {
		country = currentAddress.Country
	}

	addressLine2 := req.AddressLine2
	if addressLine2 == nil {
		addressLine2 = currentAddress.AddressLine2
	}

	latitude := req.Latitude
	if latitude == nil {
		latitude = currentAddress.Latitude
	}

	longitude := req.Longitude
	if longitude == nil {
		longitude = currentAddress.Longitude
	}

	isPrimary := currentAddress.IsPrimary
	if req.IsPrimary != nil {
		isPrimary = *req.IsPrimary
	}

	addressType := currentAddress.AddressType
	if req.AddressType != "" {
		addressType = entities.AddressType(req.AddressType)
	}

	address, err := c.addressService.UpdateAddress(
		ctx.Context(),
		id,
		addressLine1,
		city,
		state,
		postalCode,
		country,
		addressLine2,
		latitude,
		longitude,
		isPrimary,
		addressType,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   address,
	})
}

// DeleteAddress godoc
//	@Summary		Delete an address
//	@Description	Delete an address by its ID
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Address ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse	"Address deleted successfully"
//	@Failure		400	{object}	responses.ErrorResponse		"Bad request - Invalid address ID format"
//	@Failure		500	{object}	responses.ErrorResponse		"Internal server error"
//	@Router			/api/v1/addresses/{id} [delete]
//	@Security		BearerAuth
func (c *AddressController) DeleteAddress(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address ID format",
		})
	}

	err = c.addressService.DeleteAddress(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Address deleted successfully",
	})
}

// GetAddressesByAddressable godoc
//	@Summary		Get addresses by addressable entity
//	@Description	Retrieve all addresses for a specific user or organization
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			addressable_id		path		string												true	"Addressable entity ID"		format(uuid)
//	@Param			addressable_type	path		string												true	"Addressable entity type"	Enums(user, organization)
//	@Success		200					{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400					{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500					{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addressables/{addressable_type}/{addressable_id}/addresses [get]
//	@Security		BearerAuth
func (c *AddressController) GetAddressesByAddressable(ctx *fiber.Ctx) error {
	addressableIDStr := ctx.Params("addressable_id")
	addressableTypeStr := ctx.Params("addressable_type")

	addressableID, err := uuid.Parse(addressableIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable ID format",
		})
	}

	addressableType := entities.AddressableType(addressableTypeStr)
	if addressableType != entities.AddressableTypeUser && addressableType != entities.AddressableTypeOrganization {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable type",
		})
	}

	addresses, err := c.addressService.GetAddressesByAddressable(ctx.Context(), addressableID, addressableType)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}

// GetPrimaryAddressByAddressable godoc
//	@Summary		Get primary address by addressable entity
//	@Description	Retrieve the primary address for a specific user or organization
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			addressable_id		path		string												true	"Addressable entity ID"		format(uuid)
//	@Param			addressable_type	path		string												true	"Addressable entity type"	Enums(user, organization)
//	@Success		200					{object}	responses.SuccessResponse{data=entities.Address}	"Primary address found"
//	@Failure		400					{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		404					{object}	responses.ErrorResponse								"Primary address not found"
//	@Router			/api/v1/addressables/{addressable_type}/{addressable_id}/addresses/primary [get]
//	@Security		BearerAuth
func (c *AddressController) GetPrimaryAddressByAddressable(ctx *fiber.Ctx) error {
	addressableIDStr := ctx.Params("addressable_id")
	addressableTypeStr := ctx.Params("addressable_type")

	addressableID, err := uuid.Parse(addressableIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable ID format",
		})
	}

	addressableType := entities.AddressableType(addressableTypeStr)
	if addressableType != entities.AddressableTypeUser && addressableType != entities.AddressableTypeOrganization {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable type",
		})
	}

	address, err := c.addressService.GetPrimaryAddressByAddressable(ctx.Context(), addressableID, addressableType)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Primary address not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   address,
	})
}

// GetAddressesByAddressableAndType godoc
//	@Summary		Get addresses by addressable entity and type
//	@Description	Retrieve addresses for a specific user or organization filtered by address type
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			addressable_id		path		string												true	"Addressable entity ID"		format(uuid)
//	@Param			addressable_type	path		string												true	"Addressable entity type"	Enums(user, organization)
//	@Param			address_type		path		string												true	"Address type"				Enums(home, work, billing, shipping, other)
//	@Success		200					{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400					{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500					{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addressables/{addressable_type}/{addressable_id}/addresses/type/{address_type} [get]
//	@Security		BearerAuth
func (c *AddressController) GetAddressesByAddressableAndType(ctx *fiber.Ctx) error {
	addressableIDStr := ctx.Params("addressable_id")
	addressableTypeStr := ctx.Params("addressable_type")
	addressTypeStr := ctx.Params("address_type")

	addressableID, err := uuid.Parse(addressableIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable ID format",
		})
	}

	addressableType := entities.AddressableType(addressableTypeStr)
	if addressableType != entities.AddressableTypeUser && addressableType != entities.AddressableTypeOrganization {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable type",
		})
	}

	addressType := entities.AddressType(addressTypeStr)
	if addressType != entities.AddressTypeHome && addressType != entities.AddressTypeWork &&
		addressType != entities.AddressTypeBilling && addressType != entities.AddressTypeShipping &&
		addressType != entities.AddressTypeOther {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address type",
		})
	}

	addresses, err := c.addressService.GetAddressesByAddressableAndType(ctx.Context(), addressableID, addressableType, addressType)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}

// SetPrimaryAddress godoc
//	@Summary		Set address as primary
//	@Description	Set a specific address as the primary address for an addressable entity
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Address ID"	format(uuid)
//	@Param			request	body		requests.SetPrimaryAddressRequest	true	"Set primary address request"
//	@Success		200		{object}	responses.SuccessResponse			"Address set as primary successfully"
//	@Failure		400		{object}	responses.ErrorResponse				"Bad request - Invalid input data"
//	@Failure		500		{object}	responses.ErrorResponse				"Internal server error"
//	@Router			/api/v1/addresses/{id}/primary [put]
//	@Security		BearerAuth
func (c *AddressController) SetPrimaryAddress(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address ID format",
		})
	}

	var req requests.SetPrimaryAddressRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	addressableID, err := uuid.Parse(req.AddressableID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid addressable ID format",
		})
	}

	addressableType := entities.AddressableType(req.AddressableType)

	err = c.addressService.SetPrimaryAddress(ctx.Context(), id, addressableID, addressableType)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Address set as primary successfully",
	})
}

// SetAddressType godoc
//	@Summary		Set address type
//	@Description	Update the type of an address
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Address ID"	format(uuid)
//	@Param			request	body		requests.SetAddressTypeRequest	true	"Set address type request"
//	@Success		200		{object}	responses.SuccessResponse		"Address type updated successfully"
//	@Failure		400		{object}	responses.ErrorResponse			"Bad request - Invalid input data"
//	@Failure		500		{object}	responses.ErrorResponse			"Internal server error"
//	@Router			/api/v1/addresses/{id}/type [put]
//	@Security		BearerAuth
func (c *AddressController) SetAddressType(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid address ID format",
		})
	}

	var req requests.SetAddressTypeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	addressType := entities.AddressType(req.AddressType)

	err = c.addressService.SetAddressType(ctx.Context(), id, addressType)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Address type updated successfully",
	})
}

// SearchAddressesByCity godoc
//	@Summary		Search addresses by city
//	@Description	Search for addresses in a specific city with pagination
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			city	query		string												true	"City name to search for"
//	@Param			limit	query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses/search/city [get]
//	@Security		BearerAuth
func (c *AddressController) SearchAddressesByCity(ctx *fiber.Ctx) error {
	city := ctx.Query("city")
	if city == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "City parameter is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	addresses, err := c.addressService.SearchAddressesByCity(ctx.Context(), city, limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}

// SearchAddressesByState godoc
//	@Summary		Search addresses by state
//	@Description	Search for addresses in a specific state with pagination
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			state	query		string												true	"State name to search for"
//	@Param			limit	query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses/search/state [get]
//	@Security		BearerAuth
func (c *AddressController) SearchAddressesByState(ctx *fiber.Ctx) error {
	state := ctx.Query("state")
	if state == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "State parameter is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	addresses, err := c.addressService.SearchAddressesByState(ctx.Context(), state, limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}

// SearchAddressesByCountry godoc
//	@Summary		Search addresses by country
//	@Description	Search for addresses in a specific country with pagination
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			country	query		string												true	"Country name to search for"
//	@Param			limit	query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses/search/country [get]
//	@Security		BearerAuth
func (c *AddressController) SearchAddressesByCountry(ctx *fiber.Ctx) error {
	country := ctx.Query("country")
	if country == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Country parameter is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	addresses, err := c.addressService.SearchAddressesByCountry(ctx.Context(), country, limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}

// SearchAddressesByPostalCode godoc
//	@Summary		Search addresses by postal code
//	@Description	Search for addresses with a specific postal code with pagination
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			postal_code	query		string												true	"Postal code to search for"
//	@Param			limit		query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset		query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200			{object}	responses.SuccessResponse{data=[]entities.Address}	"Addresses found"
//	@Failure		400			{object}	responses.ErrorResponse								"Bad request - Invalid parameters"
//	@Failure		500			{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/addresses/search/postal-code [get]
//	@Security		BearerAuth
func (c *AddressController) SearchAddressesByPostalCode(ctx *fiber.Ctx) error {
	postalCode := ctx.Query("postal_code")
	if postalCode == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Postal code parameter is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	addresses, err := c.addressService.SearchAddressesByPostalCode(ctx.Context(), postalCode, limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   addresses,
	})
}
