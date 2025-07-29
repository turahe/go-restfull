package controllers

import (
	"net/http"
	"strconv"

	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"

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
