package controllers

import (
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
)

type RBACController struct {
	rbacService services.RBACService
}

func NewRBACController(rbacService services.RBACService) *RBACController {
	return &RBACController{
		rbacService: rbacService,
	}
}

// GetPolicy godoc
// @Summary Get all RBAC policies
// @Description Retrieve all RBAC policies from the system
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.RBACPolicyCollectionResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/policies [get]
func (c *RBACController) GetPolicy(ctx *fiber.Ctx) error {
	policies, err := c.rbacService.GetPolicy()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get policies",
		})
	}

	return ctx.JSON(responses.NewRBACPolicyCollectionResponse(policies))
}

// AddPolicy godoc
// @Summary Add a new RBAC policy
// @Description Add a new policy rule to the RBAC system
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param policy body AddPolicyRequest true "Policy to add"
// @Success 200 {object} responses.RBACPolicyResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/policies [post]
func (c *RBACController) AddPolicy(ctx *fiber.Ctx) error {
	var req AddPolicyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	err := c.rbacService.AddPolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to add policy",
		})
	}

	return ctx.JSON(responses.NewRBACPolicyResourceResponse(req.Subject, req.Object, req.Action))
}

// RemovePolicy godoc
// @Summary Remove an RBAC policy
// @Description Remove a policy rule from the RBAC system
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param policy body RemovePolicyRequest true "Policy to remove"
// @Success 200 {object} responses.RBACPolicyResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/policies [delete]
func (c *RBACController) RemovePolicy(ctx *fiber.Ctx) error {
	var req RemovePolicyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	err := c.rbacService.RemovePolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to remove policy",
		})
	}

	return ctx.JSON(responses.NewRBACPolicyResourceResponse(req.Subject, req.Object, req.Action))
}

// GetRolesForUser godoc
// @Summary Get roles for a user
// @Description Retrieve all roles assigned to a specific user
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user path string true "User ID"
// @Success 200 {object} responses.RBACRoleCollectionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/users/{user}/roles [get]
func (c *RBACController) GetRolesForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "User ID is required",
		})
	}

	roles, err := c.rbacService.GetRolesForUser(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get user roles",
		})
	}

	return ctx.JSON(responses.NewRBACRoleCollectionResponse(roles))
}

// AddRoleForUser godoc
// @Summary Add role to user
// @Description Assign a role to a specific user
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user path string true "User ID"
// @Param role path string true "Role name"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/users/{user}/roles/{role} [post]
func (c *RBACController) AddRoleForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	role := ctx.Params("role")

	if userID == "" || role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "User ID and role are required",
		})
	}

	err := c.rbacService.AddRoleForUser(userID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to add role for user",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Role added to user successfully",
	})
}

// RemoveRoleForUser godoc
// @Summary Remove role from user
// @Description Remove a role from a specific user
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user path string true "User ID"
// @Param role path string true "Role name"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/users/{user}/roles/{role} [delete]
func (c *RBACController) RemoveRoleForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	role := ctx.Params("role")

	if userID == "" || role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "User ID and role are required",
		})
	}

	err := c.rbacService.RemoveRoleForUser(userID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to remove role from user",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Role removed from user successfully",
	})
}

// GetUsersForRole godoc
// @Summary Get users for a role
// @Description Retrieve all users assigned to a specific role
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role path string true "Role name"
// @Success 200 {object} responses.RBACUserCollectionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /rbac/roles/{role}/users [get]
func (c *RBACController) GetUsersForRole(ctx *fiber.Ctx) error {
	role := ctx.Params("role")
	if role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Role is required",
		})
	}

	users, err := c.rbacService.GetUsersForRole(role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get users for role",
		})
	}

	return ctx.JSON(responses.NewRBACUserCollectionResponse(users))
}

// Request structures
type AddPolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

type RemovePolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}
