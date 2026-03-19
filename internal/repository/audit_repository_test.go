package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuditRepository_CreateImpersonation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.ImpersonationAudit{})
	repo := NewAuditRepository(db, zap.NewNop())

	a := &model.ImpersonationAudit{
		ImpersonatorID:     1,
		ImpersonatedUserID: 2,
		Reason:             "support",
		IPAddress:          "127.0.0.1",
		UserAgent:          "ua",
	}
	assert.NoError(t, repo.CreateImpersonation(ctx, a))
	assert.NotZero(t, a.ID)
}

