package model

import (
	"time"

	"gorm.io/gorm"
)

type ImpersonationAudit struct {
	ID                 uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ImpersonatorID     uint      `json:"impersonatorId" gorm:"not null;index"`
	ImpersonatedUserID uint      `json:"impersonatedUserId" gorm:"not null;index"`
	Reason             string    `json:"reason" gorm:"type:varchar(255);not null"`
	IPAddress          string    `json:"ipAddress" gorm:"type:varchar(45);not null"`
	UserAgent          string    `json:"userAgent" gorm:"type:varchar(255);not null"`
	Timestamp          time.Time `json:"timestamp" gorm:"autoCreateTime;index"`
}

func (ImpersonationAudit) TableName() string {
	return "impersonation_audits"
}

func (i *ImpersonationAudit) BeforeCreate(tx *gorm.DB) error {
	i.Timestamp = time.Now()
	return nil
}

func (i *ImpersonationAudit) BeforeUpdate(tx *gorm.DB) error {
	i.Timestamp = time.Now()
	return nil
}
