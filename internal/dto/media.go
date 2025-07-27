package dto

import (
	"github.com/google/uuid"
)

type GetMediaDTO struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	FileName string    `json:"fileName"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mimetype"`
}

type MediaRelation struct {
	MediaID      uuid.UUID `json:"mediaId"`
	MediableType string    `json:"mediableType"`
	MediableId   uuid.UUID `json:"mediableId"`
	Group        string    `json:"group"`
}

type GetMediaChildrenDTO struct {
	ID       uuid.UUID     `json:"id"`
	Name     string        `json:"name"`
	FileName string        `json:"fileName"`
	Size     int64         `json:"size"`
	MimeType string        `json:"mimetype"`
	Children []GetMediaDTO `json:"children"`
}
