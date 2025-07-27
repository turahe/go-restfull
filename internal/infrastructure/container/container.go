package container

import (
	"webapi/internal/application/ports"
	appservices "webapi/internal/application/services"
	"webapi/internal/db/rdb"
	"webapi/internal/domain/repositories"
	domainservices "webapi/internal/domain/services"
	"webapi/internal/infrastructure/adapters"
	"webapi/internal/interfaces/http/controllers"
	"webapi/internal/repository"
	"webapi/pkg/email"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	// Repositories
	UserRepository     repositories.UserRepository
	PostRepository     repositories.PostRepository
	CommentRepository  repositories.CommentRepository
	MediaRepository    repositories.MediaRepository
	RoleRepository     repositories.RoleRepository
	UserRoleRepository repositories.UserRoleRepository
	MenuRepository     repositories.MenuRepository
	MenuRoleRepository repositories.MenuRoleRepository
	TagRepository      repositories.TagRepository
	TaxonomyRepository repositories.TaxonomyRepository
	ContentRepository  repositories.ContentRepository
	SettingRepository  repository.SettingRepository
	JobRepository      repository.JobRepository

	// Application Services
	UserService     ports.UserService
	PostService     ports.PostService
	CommentService  ports.CommentService
	MediaService    ports.MediaService
	RoleService     ports.RoleService
	UserRoleService ports.UserRoleService
	MenuService     ports.MenuService
	MenuRoleService ports.MenuRoleService
	TagService      ports.TagService
	TaxonomyService ports.TaxonomyService
	ContentService  ports.ContentService
	AuthService     ports.AuthService
	JobService      ports.JobService

	// Domain Services
	EmailService      domainservices.EmailService
	PasswordService   domainservices.PasswordService
	RBACService       domainservices.RBACService
	PaginationService domainservices.PaginationService

	// Controllers
	UserController     *controllers.UserController
	PostController     *controllers.PostController
	CommentController  *controllers.CommentController
	MediaController    *controllers.MediaController
	RoleController     *controllers.RoleController
	UserRoleController *controllers.UserRoleController
	MenuController     *controllers.MenuController
	MenuRoleController *controllers.MenuRoleController
	TagController      *controllers.TagController
	TaxonomyController *controllers.TaxonomyController
	AuthController     *controllers.AuthController
	RBACController     *controllers.RBACController
	JobController      *controllers.JobController
}

func NewContainer(db *pgxpool.Pool) *Container {
	container := &Container{}

	// Get centralized Redis client
	redisClient := rdb.GetRedisClient()

	// Initialize domain services
	container.PasswordService = adapters.NewBcryptPasswordService()

	emailClient := email.NewEmailService()
	container.EmailService = adapters.NewSmtpEmailService(emailClient)

	// Initialize RBAC service
	rbacService, err := adapters.NewCasbinRBACService()
	if err != nil {
		panic(err)
	}
	container.RBACService = rbacService

	// Initialize pagination service
	container.PaginationService = domainservices.NewPaginationService()

	// Initialize repositories using existing adapters
	container.UserRepository = adapters.NewPostgresUserRepository(db, redisClient)
	container.PostRepository = adapters.NewPostgresPostRepository(db, redisClient)
	container.MediaRepository = adapters.NewPostgresMediaRepository(db, redisClient)
	container.TagRepository = adapters.NewPostgresTagRepository(db, redisClient)
	container.CommentRepository = adapters.NewPostgresCommentRepository(db, redisClient)
	container.RoleRepository = adapters.NewPostgresRoleRepository(db, redisClient)
	container.UserRoleRepository = adapters.NewPostgresUserRoleRepository(db, redisClient)
	container.MenuRepository = adapters.NewPostgresMenuRepository(db, redisClient)
	container.MenuRoleRepository = adapters.NewPostgresMenuRoleRepository(db, redisClient)
	container.TaxonomyRepository = adapters.NewPostgresTaxonomyRepository(db, redisClient)
	container.ContentRepository = adapters.NewPostgresContentRepository(db)
	container.SettingRepository = repository.NewSettingRepository(db, redisClient)
	container.JobRepository = repository.NewJobRepository(db, redisClient)

	// Initialize application services
	container.UserService = appservices.NewUserService(
		container.UserRepository,
		container.PasswordService,
		container.EmailService,
	)
	container.AuthService = appservices.NewAuthService(
		container.UserRepository,
		container.PasswordService,
		container.EmailService,
	)
	container.PostService = appservices.NewPostService(container.PostRepository)
	container.MediaService = appservices.NewMediaService(container.MediaRepository)
	container.TagService = appservices.NewTagService(container.TagRepository)
	container.CommentService = appservices.NewCommentService(container.CommentRepository)
	container.RoleService = appservices.NewRoleService(container.RoleRepository)
	container.UserRoleService = appservices.NewUserRoleService(container.UserRoleRepository)
	container.MenuService = appservices.NewMenuService(container.MenuRepository)
	container.MenuRoleService = appservices.NewMenuRoleService(container.MenuRoleRepository)
	container.TaxonomyService = appservices.NewTaxonomyService(container.TaxonomyRepository)
	container.ContentService = appservices.NewContentService(container.ContentRepository)
	container.JobService = appservices.NewJobService(container.JobRepository)

	// Initialize controllers
	container.UserController = controllers.NewUserController(container.UserService, container.PaginationService)
	container.AuthController = controllers.NewAuthController(container.AuthService)
	container.PostController = controllers.NewPostController(container.PostService)
	container.MediaController = controllers.NewMediaController(container.MediaService)
	container.TagController = controllers.NewTagController(container.TagService)
	container.CommentController = controllers.NewCommentController(container.CommentService)
	container.RoleController = controllers.NewRoleController(container.RoleService)
	container.UserRoleController = controllers.NewUserRoleController(container.UserRoleService)
	container.MenuController = controllers.NewMenuController(container.MenuService)
	container.MenuRoleController = controllers.NewMenuRoleController(container.MenuRoleService)
	container.TaxonomyController = controllers.NewTaxonomyController(container.TaxonomyService)
	container.RBACController = controllers.NewRBACController(container.RBACService)
	container.JobController = controllers.NewJobController(container.JobService)

	return container
}

// Getter methods for controllers
func (c *Container) GetUserController() *controllers.UserController {
	return c.UserController
}

func (c *Container) GetAuthController() *controllers.AuthController {
	return c.AuthController
}

func (c *Container) GetPostController() *controllers.PostController {
	return c.PostController
}

func (c *Container) GetMediaController() *controllers.MediaController {
	return c.MediaController
}

func (c *Container) GetTagController() *controllers.TagController {
	return c.TagController
}

func (c *Container) GetCommentController() *controllers.CommentController {
	return c.CommentController
}

func (c *Container) GetRoleController() *controllers.RoleController {
	return c.RoleController
}

func (c *Container) GetUserRoleController() *controllers.UserRoleController {
	return c.UserRoleController
}

func (c *Container) GetMenuController() *controllers.MenuController {
	return c.MenuController
}

func (c *Container) GetMenuRoleController() *controllers.MenuRoleController {
	return c.MenuRoleController
}

func (c *Container) GetTaxonomyController() *controllers.TaxonomyController {
	return c.TaxonomyController
}

func (c *Container) GetRBACController() *controllers.RBACController {
	return c.RBACController
}

func (c *Container) GetJobController() *controllers.JobController {
	return c.JobController
}
