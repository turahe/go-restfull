package adapters

import (
	"context"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresUserRoleRepository struct {
	repo repository.UserRoleRepository
}

func NewPostgresUserRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.UserRoleRepository {
	return &PostgresUserRoleRepository{
		repo: repository.NewUserRoleRepository(db, redisClient),
	}
}

func (r *PostgresUserRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (r *PostgresUserRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

func (r *PostgresUserRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	// The repository should return entities directly, not models
	// If the repository returns models, we need to convert them
	// For now, assuming the repository returns entities
	return r.repo.GetUserRoles(ctx, userID)
}

func (r *PostgresUserRoleRepository) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	// The repository should return entities directly, not models
	// If the repository returns models, we need to convert them
	// For now, assuming the repository returns entities
	return r.repo.GetRoleUsers(ctx, roleID, limit, offset)
}

func (r *PostgresUserRoleRepository) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	return r.repo.HasRole(ctx, userID, roleID)
}

func (r *PostgresUserRoleRepository) HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	return r.repo.HasAnyRole(ctx, userID, roleIDs)
}

func (r *PostgresUserRoleRepository) GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return r.repo.GetUserRoleIDs(ctx, userID)
}

func (r *PostgresUserRoleRepository) CountUsersByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return r.repo.CountUsersByRole(ctx, roleID)
}
