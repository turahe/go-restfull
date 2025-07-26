package container

import (
	"webapi/internal/application/ports"
	appservices "webapi/internal/application/services"
	"webapi/internal/domain/repositories"
	domainservices "webapi/internal/domain/services"
	"webapi/internal/infrastructure/adapters"
	"webapi/internal/interfaces/http/controllers"

	"github.com/jackc/pgx/v5/pgxpool"
	"webapi/pkg/email"
)

// Container holds all the dependencies
type Container struct {
	// Domain Services
	PasswordService domainservices.PasswordService
	EmailService    domainservices.EmailService

	// Repositories
	UserRepository repositories.UserRepository
	PostRepository repositories.PostRepository

	// Application Services
	UserService ports.UserService
	PostService ports.PostService

	// Controllers
	UserController *controllers.UserController
	// PostController *controllers.PostController // TODO: Implement
}

// NewContainer creates a new dependency injection container
func NewContainer(db *pgxpool.Pool) *Container {
	container := &Container{}

	// Initialize domain services
	container.PasswordService = adapters.NewBcryptPasswordService()
	
	emailClient := email.NewEmailService()
	container.EmailService = adapters.NewSmtpEmailService(emailClient)

	// Initialize repositories
	container.UserRepository = adapters.NewPostgresUserRepository(db)
	// container.PostRepository = adapters.NewPostgresPostRepository(db) // TODO: Implement

	// Initialize application services
	container.UserService = appservices.NewUserService(
		container.UserRepository,
		container.PasswordService,
		container.EmailService,
	)
	// container.PostService = appservices.NewPostService(container.PostRepository) // TODO: Implement

	// Initialize controllers
	container.UserController = controllers.NewUserController(container.UserService)
	// container.PostController = controllers.NewPostController(container.PostService) // TODO: Implement

	return container
}

// GetUserController returns the user controller
func (c *Container) GetUserController() *controllers.UserController {
	return c.UserController
}

// GetPostController returns the post controller
// func (c *Container) GetPostController() *controllers.PostController {
// 	return c.PostController
// } 