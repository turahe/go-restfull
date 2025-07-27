package routes

import (
	"webapi/internal/http/controllers/healthz"
	"webapi/internal/infrastructure/container"
	"webapi/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterHexagonalRoutes registers all routes using the Hexagonal Architecture
func RegisterHexagonalRoutes(app *fiber.App, container *container.Container) {
	// API v1 routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Comprehensive health check endpoint
	healthzHandler := healthz.NewHealthzHTTPHandler()
	app.Get("/healthz", healthzHandler.Healthz)

	// Public routes (no authentication required)
	public := v1.Group("/")

	// Auth routes (public)
	auth := public.Group("/auth")
	authController := container.GetAuthController()
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/forget-password", authController.ForgetPassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// Protected routes (require JWT + RBAC)
	rbacProtected := v1.Group("/", middleware.JWTAuth(), middleware.RBACMiddleware(container.RBACService))

	// Auth routes (protected, e.g. logout)
	authProtected := rbacProtected.Group("/auth")
	authProtected.Post("/logout", authController.Logout)

	// User routes (protected)
	users := rbacProtected.Group("/users")
	userController := container.GetUserController()
	users.Post("/", userController.CreateUser)
	users.Get("/", userController.GetUsers)
	users.Get("/:id", userController.GetUserByID)
	users.Put("/:id", userController.UpdateUser)
	users.Delete("/:id", userController.DeleteUser)
	users.Put("/:id/password", userController.ChangePassword)
	users.Get("/profile", userController.GetUserByID)
	users.Put("/profile", userController.UpdateUser)

	// Post routes (protected)
	postController := container.GetPostController()
	posts := rbacProtected.Group("/posts")
	posts.Post("/", postController.CreatePost)
	posts.Get("/", postController.GetPosts)
	posts.Get("/:id", postController.GetPostByID)
	posts.Get("/slug/:slug", postController.GetPostBySlug)
	posts.Get("/author/:authorID", postController.GetPostsByAuthor)
	posts.Put("/:id", postController.UpdatePost)
	posts.Delete("/:id", postController.DeletePost)
	posts.Put("/:id/publish", postController.PublishPost)
	posts.Put("/:id/unpublish", postController.UnpublishPost)

	// Menu routes (protected)
	menuController := container.GetMenuController()
	menus := rbacProtected.Group("/menus")
	menus.Post("/", menuController.CreateMenu)
	menus.Get("/", menuController.GetMenus)
	menus.Get("/root", menuController.GetRootMenus)
	menus.Get("/hierarchy", menuController.GetMenuHierarchy)
	menus.Get("/search", menuController.SearchMenus)
	menus.Get("/slug/:slug", menuController.GetMenuBySlug)
	menus.Get("/:id", menuController.GetMenuByID)
	menus.Put("/:id", menuController.UpdateMenu)
	menus.Delete("/:id", menuController.DeleteMenu)
	menus.Patch("/:id/activate", menuController.ActivateMenu)
	menus.Patch("/:id/deactivate", menuController.DeactivateMenu)
	menus.Patch("/:id/show", menuController.ShowMenu)
	menus.Patch("/:id/hide", menuController.HideMenu)

	// Menu-Role routes (protected)
	menuRoleController := container.GetMenuRoleController()
	menus.Post("/:menu_id/roles/:role_id", menuRoleController.AssignRoleToMenu)
	menus.Delete("/:menu_id/roles/:role_id", menuRoleController.RemoveRoleFromMenu)
	menus.Get("/:menu_id/roles", menuRoleController.GetMenuRoles)
	menus.Get("/:menu_id/roles/:role_id/check", menuRoleController.HasRole)

	// User-Menu routes (protected)
	users.Get("/:user_id/menus", menuController.GetUserMenus)

	// Role-Menu routes (protected)
	roles := rbacProtected.Group("/roles")
	roles.Get("/:role_id/menus", menuRoleController.GetRoleMenus)
	roles.Get("/:role_id/menus/count", menuRoleController.GetMenuRoleCount)

	// Taxonomy routes (protected)
	taxonomyController := container.GetTaxonomyController()
	taxonomies := rbacProtected.Group("/taxonomies")
	taxonomies.Post("/", taxonomyController.CreateTaxonomy)
	taxonomies.Get("/", taxonomyController.GetTaxonomies)
	taxonomies.Get("/root", taxonomyController.GetRootTaxonomies)
	taxonomies.Get("/hierarchy", taxonomyController.GetTaxonomyHierarchy)
	taxonomies.Get("/search", taxonomyController.SearchTaxonomies)
	taxonomies.Get("/slug/:slug", taxonomyController.GetTaxonomyBySlug)
	taxonomies.Get("/:id", taxonomyController.GetTaxonomyByID)
	taxonomies.Put("/:id", taxonomyController.UpdateTaxonomy)
	taxonomies.Delete("/:id", taxonomyController.DeleteTaxonomy)
	taxonomies.Get("/:id/children", taxonomyController.GetTaxonomyChildren)
	taxonomies.Get("/:id/descendants", taxonomyController.GetTaxonomyDescendants)
	taxonomies.Get("/:id/ancestors", taxonomyController.GetTaxonomyAncestors)
	taxonomies.Get("/:id/siblings", taxonomyController.GetTaxonomySiblings)

	// RBAC Management routes (admin only)
	rbacController := container.GetRBACController()
	rbac := rbacProtected.Group("/rbac")
	rbac.Get("/policies", rbacController.GetPolicy)
	rbac.Post("/policies", rbacController.AddPolicy)
	rbac.Delete("/policies", rbacController.RemovePolicy)
	rbac.Get("/users/:user_id/roles", rbacController.GetRolesForUser)
	rbac.Post("/users/:user_id/roles", rbacController.AddRoleForUser)
	rbac.Delete("/users/:user_id/roles", rbacController.RemoveRoleForUser)
	rbac.Get("/roles/:role_id/users", rbacController.GetUsersForRole)

	// Job Management routes (admin only)
	jobController := container.GetJobController()
	jobs := rbacProtected.Group("/jobs")
	jobs.Get("/", jobController.GetJobs)
	jobs.Get("/:id", jobController.GetJob)
	jobs.Get("/failed", jobController.GetFailedJobs)
	jobs.Post("/:id/retry", jobController.RetryFailedJob)
	jobs.Delete("/failed/:id", jobController.RemoveFailedJob)
	jobs.Post("/process", jobController.ProcessJobs)
	jobs.Post("/reset", jobController.ResetProcessingJobs)

	// TODO: Add other routes as they are implemented in Hexagonal Architecture
	// - Media routes
	// - Setting routes
	// - Tag routes
}
