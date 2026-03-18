package repository

import (
	"context"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) CreateImpersonation(ctx context.Context, a *model.ImpersonationAudit) error {
	return r.db.WithContext(ctx).Create(a).Error
}

