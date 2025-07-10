package model

import "time"

type Comment struct {
	ID             string    `json:"id"`
	ModelType      string    `json:"model_type"`
	ModelID        string    `json:"model_id"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	ParentID       string    `json:"parent_id"`
	RecordLeft     int64     `json:"record_left"`
	RecordRight    int64     `json:"record_right"`
	RecordOrdering int64     `json:"record_ordering"`
	CreatedBy      string    `json:"created_by"`
	UpdatedBy      string    `json:"updated_by"`
	DeletedBy      string    `json:"deleted_by"`
	DeletedAt      time.Time `json:"deleted_at"`
}
