package domain

import "time"

// Category is the domain representation of a row in categories (nested set).
type Category struct {
	ID       uint
	Name     string
	Slug     string
	ParentID *uint
	Lft      int
	Rgt      int
	Depth    int

	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
