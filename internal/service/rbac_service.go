package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/rbac"

	"gorm.io/gorm"
)

type RBACService struct {
	e  *rbac.Enforcer
	db *gorm.DB
}

func NewRBACService(e *rbac.Enforcer, db *gorm.DB) *RBACService {
	return &RBACService{e: e, db: db}
}

func (s *RBACService) RolesForUser(_ context.Context, userID uint) ([]string, error) {
	return s.e.GetRolesForUser(fmt.Sprintf("%d", userID))
}

// PermissionsForUser returns implicit permissions as "obj:act" strings.
func (s *RBACService) PermissionsForUser(_ context.Context, userID uint) ([]string, error) {
	perms, err := s.e.GetImplicitPermissionsForUser(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		// p = [sub obj act]
		if len(p) < 3 {
			continue
		}
		out = append(out, strings.TrimSpace(p[1])+":"+strings.TrimSpace(p[2]))
	}
	return out, nil
}

func (s *RBACService) Enforce(_ context.Context, userID uint, obj string, act string) (bool, error) {
	return s.e.Enforce(fmt.Sprintf("%d", userID), obj, act)
}

// Admin helpers
func (s *RBACService) AssignRole(ctx context.Context, userID uint, role string) (bool, error) {
	role = strings.TrimSpace(role)
	if role == "" {
		return false, errors.New("role is required")
	}

	// Persist to RBAC tables (roles, user_roles) if DB is available.
	if s.db != nil {
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			r := model.Role{Name: role}
			if err := tx.Where("name = ?", role).FirstOrCreate(&r).Error; err != nil {
				return err
			}
			ur := model.UserRole{UserID: userID, RoleID: r.ID}
			return tx.Where("user_id = ? AND role_id = ?", userID, r.ID).FirstOrCreate(&ur).Error
		}); err != nil {
			return false, err
		}
	}

	// Also persist to Casbin grouping policy for enforcement.
	return s.e.AddRoleForUser(fmt.Sprintf("%d", userID), role)
}

func (s *RBACService) AddPermissionToRole(_ context.Context, role, obj, act string) (bool, error) {
	return s.e.AddPolicy(role, obj, act)
}

