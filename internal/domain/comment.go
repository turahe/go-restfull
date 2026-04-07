package domain

import "time"

// Comment is the domain representation of a threaded comment on a post (nested set scoped by post_id).
type Comment struct {
	ID       uint
	PostID   uint
	ParentID *uint
	Lft      int
	Rgt      int
	Depth    int

	UserID  uint
	Content string

	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
