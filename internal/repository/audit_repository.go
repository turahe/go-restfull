package repository

import (
	"context"

	"go-rest/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewAuditRepository(db *gorm.DB, log *zap.Logger) *AuditRepository {
	return &AuditRepository{db: db, log: log}
}

func (r *AuditRepository) CreateImpersonation(ctx context.Context, a *model.ImpersonationAudit) error {
	err := r.db.WithContext(ctx).Create(a).Error
	if err != nil {
		r.log.Error("failed to create impersonation audit", zap.Error(err))
		return err
	}
	return nil
}
