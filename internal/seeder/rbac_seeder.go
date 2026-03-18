package seeder

import (
	"context"
	"errors"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/rbac"

	"gorm.io/gorm"
)

type PermissionSeed struct {
	Role string
	Obj  string
	Act  string
	Desc string
}

func SeedDefaultRBAC(ctx context.Context, db *gorm.DB, enf *rbac.Enforcer) error {
	if db == nil {
		return errors.New("db is required")
	}
	if enf == nil {
		return errors.New("enforcer is required")
	}

	roles := []string{"admin", "support", "user"}
	seeds := []PermissionSeed{
		// Admin (full access)
		{Role: "admin", Obj: "/api/v1/*", Act: ".*", Desc: "Full API access"},

		// Support (everything except RBAC admin endpoints)
		{Role: "support", Obj: "/api/v1/auth/impersonate", Act: "POST", Desc: "Impersonate users"},
		{Role: "support", Obj: "/api/v1/auth/password/change", Act: "POST", Desc: "Change password"},
		{Role: "support", Obj: "/api/v1/auth/email/change", Act: "POST", Desc: "Change email"},
		{Role: "support", Obj: "/api/v1/posts*", Act: "(GET|POST|PUT|DELETE)", Desc: "Manage posts"},
		{Role: "support", Obj: "/api/v1/categories*", Act: "(GET|POST|PUT|DELETE)", Desc: "Manage categories"},
		{Role: "support", Obj: "/api/v1/tags*", Act: "(GET|POST|PUT|DELETE)", Desc: "Manage tags"},
		{Role: "support", Obj: "/api/v1/media*", Act: "(GET|POST|DELETE)", Desc: "Manage media"},
		{Role: "support", Obj: "/api/v1/posts/*/comments", Act: "POST", Desc: "Create comments"},
		{Role: "support", Obj: "/api/v1/posts/*/comments", Act: "GET", Desc: "List comments"},

		// User (basic CRUD; ownership checks are handled elsewhere)
		{Role: "user", Obj: "/api/v1/posts", Act: "GET", Desc: "List posts"},
		{Role: "user", Obj: "/api/v1/posts/slug/*", Act: "GET", Desc: "Get post by slug"},
		{Role: "user", Obj: "/api/v1/categories", Act: "GET", Desc: "List categories"},
		{Role: "user", Obj: "/api/v1/categories/*", Act: "GET", Desc: "Get category by slug"},
		{Role: "user", Obj: "/api/v1/tags", Act: "GET", Desc: "List tags"},
		{Role: "user", Obj: "/api/v1/tags/*", Act: "GET", Desc: "Get tag by slug"},
		{Role: "user", Obj: "/api/v1/media*", Act: "(GET|POST|DELETE)", Desc: "Manage media"},
		{Role: "user", Obj: "/api/v1/auth/password/change", Act: "POST", Desc: "Change password"},
		{Role: "user", Obj: "/api/v1/auth/email/change", Act: "POST", Desc: "Change email"},
		{Role: "user", Obj: "/api/v1/posts", Act: "POST", Desc: "Create post"},
		{Role: "user", Obj: "/api/v1/posts/*", Act: "PUT", Desc: "Update post"},
		{Role: "user", Obj: "/api/v1/posts/*", Act: "DELETE", Desc: "Delete post"},
		{Role: "user", Obj: "/api/v1/posts/*/comments", Act: "POST", Desc: "Create comment"},
		{Role: "user", Obj: "/api/v1/posts/*/comments", Act: "GET", Desc: "List comments"},
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Seed Role + Permission tables (optional but useful for admin UI).
		roleByName := map[string]model.Role{}
		for _, r := range roles {
			name := strings.TrimSpace(r)
			if name == "" {
				continue
			}
			role := model.Role{Name: name}
			if err := tx.Where("name = ?", name).FirstOrCreate(&role).Error; err != nil {
				return err
			}
			roleByName[name] = role
		}

		permByKey := map[string]model.Permission{}
		for _, s := range seeds {
			key := strings.TrimSpace(s.Obj) + ":" + strings.TrimSpace(s.Act)
			p := model.Permission{Key: key, Desc: s.Desc}
			if err := tx.Where("`key` = ?", key).FirstOrCreate(&p).Error; err != nil {
				return err
			}
			if p.Desc == "" && s.Desc != "" {
				_ = tx.Model(&model.Permission{}).Where("id = ?", p.ID).Update("desc", s.Desc).Error
			}
			permByKey[key] = p
		}

		for _, s := range seeds {
			role, ok := roleByName[s.Role]
			if !ok {
				continue
			}
			key := strings.TrimSpace(s.Obj) + ":" + strings.TrimSpace(s.Act)
			perm, ok := permByKey[key]
			if !ok {
				continue
			}

			rp := model.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
			if err := tx.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).FirstOrCreate(&rp).Error; err != nil {
				return err
			}
		}

		// Seed Casbin policies.
		// Note: matcher uses keyMatch2(obj) and regexMatch(act), so patterns are allowed.
		enf.EnableAutoSave(false)
		for _, s := range seeds {
			_, _ = enf.AddPolicy(s.Role, s.Obj, s.Act)
		}
		if err := enf.SavePolicy(); err != nil {
			return err
		}
		enf.EnableAutoSave(true)

		return nil
	})
}

