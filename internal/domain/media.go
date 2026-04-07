package domain

import "time"

// Media is the domain representation of user-scoped media (nested set scoped by user_id).
type Media struct {
	ID       uint
	UserID   uint
	Name     string
	ParentID *uint
	Lft      int
	Rgt      int
	Depth    int

	MediaType    string
	OriginalName string
	MimeType     string
	Size         int64
	StoragePath  string

	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
