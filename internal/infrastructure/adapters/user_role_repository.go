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
	roleModels, err := r.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	var roles []*entities.Role
	for _, roleModel := range roleModels {
		roleID, _ := uuid.Parse(roleModel.ID)
		role := &entities.Role{
			ID:          roleID,
			Name:        roleModel.Name,
			Slug:        roleModel.Slug,
			Description: roleModel.Description,
			IsActive:    roleModel.IsActive,
			CreatedAt:   roleModel.CreatedAt,
			UpdatedAt:   roleModel.UpdatedAt,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *PostgresUserRoleRepository) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	userModels, err := r.repo.GetRoleUsers(ctx, roleID, limit, offset)
	if err != nil {
		return nil, err
	}

	var users []*entities.User
	for _, userModel := range userModels {
		user := &entities.User{
			ID:        userModel.ID,
			UserName:  userModel.UserName,
			Email:     userModel.Email,
			Phone:     userModel.Phone,
			Password:  userModel.Password,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		}

		// Handle email verification
		if !userModel.EmailVerified.IsZero() {
			user.EmailVerifiedAt = &userModel.EmailVerified
		}

		// Handle phone verification
		if !userModel.PhoneVerified.IsZero() {
			user.PhoneVerifiedAt = &userModel.PhoneVerified
		}

		// Handle soft delete
		if !userModel.DeletedAt.IsZero() {
			user.DeletedAt = &userModel.DeletedAt
		}

		users = append(users, user)
	}

	return users, nil
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
