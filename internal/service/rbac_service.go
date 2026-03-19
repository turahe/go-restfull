package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/rbac"

	"github.com/casbin/casbin/v3/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RBACService struct {
	e   *rbac.Enforcer
	db  *gorm.DB
	log *zap.Logger
}

func NewRBACService(e *rbac.Enforcer, db *gorm.DB, log *zap.Logger) *RBACService {
	return &RBACService{e: e, db: db, log: log}
}

func (s *RBACService) RolesForUser(ctx context.Context, userID uint) ([]string, error) {
	// Prefer DB roles table (source of truth for user_roles).
	if s.db != nil {
		var names []string
		err := s.db.WithContext(ctx).
			Table("user_roles").
			Select("roles.name").
			Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.deleted_at IS NULL").
			Where("user_roles.user_id = ?", userID).
			Order("roles.id asc").
			Scan(&names).Error
		if err != nil {
			s.log.Error("failed to get roles for user", zap.Error(err))
			return nil, err
		}
		return names, nil
	}

	// Fallback to Casbin grouping policies.
	roles, err := s.e.GetRolesForUser(fmt.Sprintf("%d", userID))
	if err != nil {
		s.log.Error("failed to get roles for user", zap.Error(err))
		return nil, err
	}
	return roles, nil
}

// PermissionsForUser returns implicit permissions as "obj:act" strings.
func (s *RBACService) PermissionsForUser(ctx context.Context, userID uint) ([]string, error) {
	// Prefer DB permissions tables (user_roles -> role_permissions -> permissions).
	if s.db != nil {
		var keys []string
		err := s.db.WithContext(ctx).
			Table("user_roles").
			Select("permissions.key").
			Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.deleted_at IS NULL").
			Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id AND permissions.deleted_at IS NULL").
			Where("user_roles.user_id = ?", userID).
			Order("permissions.id asc").
			Scan(&keys).Error
		if err != nil {
			s.log.Error("failed to get permissions for user", zap.Error(err))
			return nil, err
		}
		out := make([]string, 0, len(keys))
		for _, k := range keys {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			out = append(out, k)
		}
		return out, nil
	}

	// Fallback to Casbin (implicit permissions).
	perms, err := s.e.GetImplicitPermissionsForUser(fmt.Sprintf("%d", userID))
	if err != nil {
		s.log.Error("failed to get implicit permissions for user", zap.Error(err))
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

func (s *RBACService) Enforce(ctx context.Context, userID uint, obj string, act string) (bool, error) {
	// Prefer DB-based enforcement: evaluate permission patterns (keyMatch2 + regexMatch).
	// Permission keys are stored as "obj:act" where obj may contain keyMatch2 wildcards
	// and act may be a regex (same semantics as the Casbin matcher).
	if s.db != nil {
		keys, err := s.PermissionsForUser(ctx, userID)
		if err != nil {
			s.log.Error("failed to get permissions for user", zap.Error(err))
			return false, err
		}
		for _, k := range keys {
			parts := strings.SplitN(k, ":", 2)
			if len(parts) != 2 {
				continue
			}
			objPat := strings.TrimSpace(parts[0])
			actPat := strings.TrimSpace(parts[1])
			if objPat == "" || actPat == "" {
				continue
			}
			if util.KeyMatch2(obj, objPat) && util.RegexMatch(act, actPat) {
				return true, nil
			}
		}
		return false, nil
	}

	// Fallback to Casbin.
	allowed, err := s.e.Enforce(fmt.Sprintf("%d", userID), obj, act)
	if err != nil {
		s.log.Error("failed to enforce", zap.Error(err))
		return false, err
	}
	return allowed, nil
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
				s.log.Error("failed to create role", zap.Error(err))
				return err
			}
			ur := model.UserRole{UserID: userID, RoleID: r.ID}
			if err := tx.Where("user_id = ? AND role_id = ?", userID, r.ID).FirstOrCreate(&ur).Error; err != nil {
				s.log.Error("failed to create user role", zap.Error(err))
				return err
			}
			return nil
		}); err != nil {
			s.log.Error("failed to assign role", zap.Error(err))
			return false, err
		}
		return true, nil
	}

	// Also persist to Casbin grouping policy for enforcement.
	added, err := s.e.AddRoleForUser(fmt.Sprintf("%d", userID), role)
	if err != nil {
		s.log.Error("failed to add role for user", zap.Error(err))
		return false, err
	}
	return added, nil
}

func (s *RBACService) AddPermissionToRole(ctx context.Context, role, obj, act string) (bool, error) {
	role = strings.TrimSpace(role)
	obj = strings.TrimSpace(obj)
	act = strings.TrimSpace(act)
	if role == "" || obj == "" || act == "" {
		return false, errors.New("role, obj, act are required")
	}

	// Persist to RBAC tables (roles, permissions, role_permissions) if DB is available.
	if s.db != nil {
		key := obj + ":" + act
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			r := model.Role{Name: role}
			if err := tx.Where("name = ?", role).FirstOrCreate(&r).Error; err != nil {
				s.log.Error("failed to create role", zap.Error(err))
				return err
			}
			p := model.Permission{Key: key}
			if err := tx.Where("`key` = ?", key).FirstOrCreate(&p).Error; err != nil {
				s.log.Error("failed to create permission", zap.Error(err))
				return err
			}
			rp := model.RolePermission{RoleID: r.ID, PermissionID: p.ID}
			if err := tx.Where("role_id = ? AND permission_id = ?", r.ID, p.ID).FirstOrCreate(&rp).Error; err != nil {
				s.log.Error("failed to create role permission", zap.Error(err))
				return err
			}
			return nil
		}); err != nil {
			s.log.Error("failed to add permission to role", zap.Error(err))
			return false, err
		}
	}

	// Also persist to Casbin policy for enforcement.
	added, err := s.e.AddPolicy(role, obj, act)
	if err != nil {
		s.log.Error("failed to add policy", zap.Error(err))
		return false, err
	}
	return added, nil
}
