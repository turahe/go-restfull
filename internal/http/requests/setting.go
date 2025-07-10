package requests

import (
	"github.com/google/uuid"
)

type CreateSettingRequest struct {
	ModelType string    `json:"modelType"`
	ModelId   uuid.UUID `json:"modelId"`
	Key       string    `json:"key" validate:"required"`
	Value     string    `json:"value" validate:"required"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
}

type UpdateSettingRequest struct {
	Key       string `json:"key" validate:"required"`
	Value     string `json:"value" validate:"required"`
	UpdatedBy string `json:"updated_by"`
}

type GetSettingByKeyRequest struct {
	Key string `json:"key" validate:"required"`
}
