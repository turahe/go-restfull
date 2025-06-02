package requests

import "github.com/google/uuid"

type CreateMediaRequest struct {
	//Name        string `json:"name" validate:"required,min=3,max=32"`
	//Description string `json:"description" validate:"required,min=3,max=255"`
	//File     string    `json:"file" validate:"required,min=3,max=255"`
	ParentID uuid.UUID `json:"parent_id"`
}
