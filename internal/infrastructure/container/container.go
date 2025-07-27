package container

import (
	"webapi/internal/application/ports"
	appservices "webapi/internal/application/services"
	"webapi/internal/domain/repositories"
	domainservices "webapi/internal/domain/services"
	"webapi/internal/infrastructure/adapters"
	"webapi/internal/interfaces/http/controllers"
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
	SettingRepository  adapters.SettingRepository
	JobRepository      adapters.JobRepository

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

	// Domain Services
	EmailService    domainservices.EmailService
	PasswordService domainservices.PasswordService
	RBACService     domainservices.RBACService

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
}

func NewContainer(db interface{}) *Container {
	// Initialize repositories
	userRepo := adapters.NewUserRepository(db)
	postRepo := adapters.NewPostRepository(db)
	commentRepo := adapters.NewCommentRepository(db)
	mediaRepo := adapters.NewMediaRepository(db)
	roleRepo := adapters.NewRoleRepository(db)
	userRoleRepo := adapters.NewUserRoleRepository(db)
	menuRepo := adapters.NewMenuRepository(db)
	menuRoleRepo := adapters.NewMenuRoleRepository(db)
	tagRepo := adapters.NewTagRepository(db)
	taxonomyRepo := adapters.NewTaxonomyRepository(db)
	contentRepo := adapters.NewContentRepository(db)
	settingRepo := adapters.NewSettingRepository(db)
	jobRepo := adapters.NewJobRepository(db)

	// Initialize domain services
	emailService := adapters.NewEmailService()
	passwordService := adapters.NewPasswordService()

	// Initialize RBAC service
	rbacService, err := adapters.NewCasbinRBACService()
	if err != nil {
		panic(err)
	}

	// Initialize application services
	userService := appservices.NewUserService(userRepo, passwordService, emailService)
	postService := appservices.NewPostService(postRepo)
	commentService := appservices.NewCommentService(commentRepo)
	mediaService := appservices.NewMediaService(mediaRepo)
	roleService := appservices.NewRoleService(roleRepo)
	userRoleService := appservices.NewUserRoleService(userRoleRepo)
	menuService := appservices.NewMenuService(menuRepo)
	menuRoleService := appservices.NewMenuRoleService(menuRoleRepo)
	tagService := appservices.NewTagService(tagRepo)
	taxonomyService := appservices.NewTaxonomyService(taxonomyRepo)
	contentService := appservices.NewContentService(contentRepo)
	authService := appservices.NewAuthService(userRepo, passwordService, emailService)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	postController := controllers.NewPostController(postService)
	commentController := controllers.NewCommentController(commentService)
	mediaController := controllers.NewMediaController(mediaService)
	roleController := controllers.NewRoleController(roleService)
	userRoleController := controllers.NewUserRoleController(userRoleService)
	menuController := controllers.NewMenuController(menuService)
	menuRoleController := controllers.NewMenuRoleController(menuRoleService)
	tagController := controllers.NewTagController(tagService)
	taxonomyController := controllers.NewTaxonomyController(taxonomyService)
	authController := controllers.NewAuthController(authService)

	return &Container{
		// Repositories
		UserRepository:     userRepo,
		PostRepository:     postRepo,
		CommentRepository:  commentRepo,
		MediaRepository:    mediaRepo,
		RoleRepository:     roleRepo,
		UserRoleRepository: userRoleRepo,
		MenuRepository:     menuRepo,
		MenuRoleRepository: menuRoleRepo,
		TagRepository:      tagRepo,
		TaxonomyRepository: taxonomyRepo,
		ContentRepository:  contentRepo,
		SettingRepository:  settingRepo,
		JobRepository:      jobRepo,

		// Application Services
		UserService:     userService,
		PostService:     postService,
		CommentService:  commentService,
		MediaService:    mediaService,
		RoleService:     roleService,
		UserRoleService: userRoleService,
		MenuService:     menuService,
		MenuRoleService: menuRoleService,
		TagService:      tagService,
		TaxonomyService: taxonomyService,
		ContentService:  contentService,
		AuthService:     authService,

		// Domain Services
		EmailService:    emailService,
		PasswordService: passwordService,
		RBACService:     rbacService,

		// Controllers
		UserController:     userController,
		PostController:     postController,
		CommentController:  commentController,
		MediaController:    mediaController,
		RoleController:     roleController,
		UserRoleController: userRoleController,
		MenuController:     menuController,
		MenuRoleController: menuRoleController,
		TagController:      tagController,
		TaxonomyController: taxonomyController,
		AuthController:     authController,
	}
}
