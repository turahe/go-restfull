package request

type UserListRequest struct {
	PageRequest
	SearchRequest
	Name  string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
	Email string `form:"email" json:"email" binding:"omitempty,email,max=190"`
}
