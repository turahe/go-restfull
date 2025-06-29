package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateSettingDTO struct {
	ModelType string    `json:"modelType"`
	ModelId   uuid.UUID `json:"modelId"`
	Key       string    `json:"key" validate:"required"`
	Value     string    `json:"value" validate:"required"`
}

type UpdateSettingDTO struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type SettingDTO struct {
	ID        uuid.UUID `json:"id"`
	ModelType string    `json:"modelType"`
	ModelId   uuid.UUID `json:"modelId"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
} 