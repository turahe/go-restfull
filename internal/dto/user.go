package dto

import (
	"github.com/google/uuid"
)

type GetUserDTO struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"username"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
}
