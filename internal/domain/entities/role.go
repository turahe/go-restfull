package entities

// Canonical role names stored in `roles.name`, assigned via RBAC, and surfaced on JWT claims.
// Keep aligned with `internal/seeder/rbac_seeder.go` seed data.
const (
	RoleAdmin   = "admin"
	RoleSupport = "support"
	RoleUser    = "user"
)
