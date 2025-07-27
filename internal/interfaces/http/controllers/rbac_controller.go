package controllers

import (
	"webapi/internal/domain/services"
	"webapi/internal/http/response"

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
// @Success 200 {object} response.CommonResponse{data=[][]string}
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/policies [get]
func (c *RBACController) GetPolicy(ctx *fiber.Ctx) error {
	policies, err := c.rbacService.GetPolicy()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get policies",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Policies retrieved successfully",
		Data:            policies,
	})
}

// AddPolicy godoc
// @Summary Add a new RBAC policy
// @Description Add a new policy rule to the RBAC system
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param policy body AddPolicyRequest true "Policy to add"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/policies [post]
func (c *RBACController) AddPolicy(ctx *fiber.Ctx) error {
	var req AddPolicyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
		})
	}

	err := c.rbacService.AddPolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to add policy",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Policy added successfully",
	})
}

// RemovePolicy godoc
// @Summary Remove an RBAC policy
// @Description Remove a policy rule from the RBAC system
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param policy body RemovePolicyRequest true "Policy to remove"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/policies [delete]
func (c *RBACController) RemovePolicy(ctx *fiber.Ctx) error {
	var req RemovePolicyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
		})
	}

	err := c.rbacService.RemovePolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to remove policy",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Policy removed successfully",
	})
}

// GetRolesForUser godoc
// @Summary Get roles for a user
// @Description Retrieve all roles assigned to a specific user
// @Tags rbac
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user path string true "User ID"
// @Success 200 {object} response.CommonResponse{data=[]string}
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/users/{user}/roles [get]
func (c *RBACController) GetRolesForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "User ID is required",
		})
	}

	roles, err := c.rbacService.GetRolesForUser(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get user roles",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "User roles retrieved successfully",
		Data:            roles,
	})
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
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/users/{user}/roles/{role} [post]
func (c *RBACController) AddRoleForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	role := ctx.Params("role")

	if userID == "" || role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "User ID and role are required",
		})
	}

	err := c.rbacService.AddRoleForUser(userID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to add role for user",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role added to user successfully",
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
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/users/{user}/roles/{role} [delete]
func (c *RBACController) RemoveRoleForUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("user")
	role := ctx.Params("role")

	if userID == "" || role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "User ID and role are required",
		})
	}

	err := c.rbacService.RemoveRoleForUser(userID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to remove role from user",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role removed from user successfully",
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
// @Success 200 {object} response.CommonResponse{data=[]string}
// @Failure 500 {object} response.CommonResponse
// @Router /rbac/roles/{role}/users [get]
func (c *RBACController) GetUsersForRole(ctx *fiber.Ctx) error {
	role := ctx.Params("role")
	if role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Role is required",
		})
	}

	users, err := c.rbacService.GetUsersForRole(role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get users for role",
		})
	}

	return ctx.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Users for role retrieved successfully",
		Data:            users,
	})
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
