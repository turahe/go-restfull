package handlers

import (
	"context"

	"github.com/turahe/go-restfull/internal/application/commands"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
)

// CommandHandler represents a generic command handler interface
type CommandHandler[T any] interface {
	Handle(ctx context.Context, cmd T) error
}

// CommandResult represents a command result with data
type CommandResult[T any] struct {
	Data  T
	Error error
}

// CommandHandlerWithResult represents a command handler that returns data
type CommandHandlerWithResult[TCmd any, TResult any] interface {
	Handle(ctx context.Context, cmd TCmd) (TResult, error)
}

// UserCommandHandlers defines all user command handlers
type UserCommandHandlers struct {
	CreateUser        CommandHandlerWithResult[commands.CreateUserCommand, *aggregates.UserAggregate]
	UpdateUser        CommandHandler[commands.UpdateUserCommand]
	ChangePassword    CommandHandler[commands.ChangePasswordCommand]
	VerifyEmail       CommandHandler[commands.VerifyEmailCommand]
	VerifyPhone       CommandHandler[commands.VerifyPhoneCommand]
	AssignRole        CommandHandler[commands.AssignRoleCommand]
	RemoveRole        CommandHandler[commands.RemoveRoleCommand]
	UpdateUserProfile CommandHandler[commands.UpdateUserProfileCommand]
	DeleteUser        CommandHandler[commands.DeleteUserCommand]
}

// CreateUserCommandHandler handles user creation commands
type CreateUserCommandHandler interface {
	CommandHandlerWithResult[commands.CreateUserCommand, *aggregates.UserAggregate]
}

// UpdateUserCommandHandler handles user update commands
type UpdateUserCommandHandler interface {
	CommandHandler[commands.UpdateUserCommand]
}

// ChangePasswordCommandHandler handles password change commands
type ChangePasswordCommandHandler interface {
	CommandHandler[commands.ChangePasswordCommand]
}

// VerifyEmailCommandHandler handles email verification commands
type VerifyEmailCommandHandler interface {
	CommandHandler[commands.VerifyEmailCommand]
}

// VerifyPhoneCommandHandler handles phone verification commands
type VerifyPhoneCommandHandler interface {
	CommandHandler[commands.VerifyPhoneCommand]
}

// AssignRoleCommandHandler handles role assignment commands
type AssignRoleCommandHandler interface {
	CommandHandler[commands.AssignRoleCommand]
}

// RemoveRoleCommandHandler handles role removal commands
type RemoveRoleCommandHandler interface {
	CommandHandler[commands.RemoveRoleCommand]
}

// UpdateUserProfileCommandHandler handles user profile update commands
type UpdateUserProfileCommandHandler interface {
	CommandHandler[commands.UpdateUserProfileCommand]
}

// DeleteUserCommandHandler handles user deletion commands
type DeleteUserCommandHandler interface {
	CommandHandler[commands.DeleteUserCommand]
}