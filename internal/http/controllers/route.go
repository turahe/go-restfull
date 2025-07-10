package httpapi

import (
	"webapi/internal/app/media"
	"webapi/internal/app/post"
	"webapi/internal/app/queue"
	"webapi/internal/app/setting"
	"webapi/internal/app/tag"
	"webapi/internal/app/user"
	httTag "webapi/internal/http/controllers/tag"
	"webapi/internal/repository"
	"webapi/internal/router/middleware"

	"github.com/gofiber/fiber/v2"

	httpAuth "webapi/internal/http/controllers/auth"
	httpHealthz "webapi/internal/http/controllers/healthz"
	httpMedia "webapi/internal/http/controllers/media"
	httpMiscellaneous "webapi/internal/http/controllers/miscellaneous"
	httpPost "webapi/internal/http/controllers/post"
	httpQueue "webapi/internal/http/controllers/queue"
	httpSetting "webapi/internal/http/controllers/setting"
	httpUser "webapi/internal/http/controllers/user"
)

// ====================================================
// =================== DEFINE ROUTE ===================
// ====================================================
var repo *repository.Repository

func RegisterRoute(r *fiber.App) {
	repo = repository.NewRepository()
	api := r.Group("/api")
	v1 := api.Group("/v1")

	// Healthz API
	healthAPI := api.Group("/healthz")
	healthHandler := httpHealthz.NewHealthzHTTPHandler()
	healthAPI.Get("/", healthHandler.Healthz)

	// Initialize apps
	userApp := user.NewUserApp(repo)
	mediaApp := media.NewMediaApp(repo)
	settingApp := setting.NewSettingApp(repo)
	queueApp := queue.NewQueueApp(repo)
	tagApp := tag.NewTagApp(repo)
	postApp := post.NewPostApp(repo)

	// Auth API (public routes)
	authApi := v1.Group("/auth")
	authHandler := httpAuth.NewAuthHttpHandler(userApp)
	authApi.Post("/register", authHandler.Register)
	authApi.Post("/login", authHandler.Login)
	authApi.Post("/refresh", authHandler.RefreshToken)
	authApi.Post("/forget-password", authHandler.ForgetPassword)
	authApi.Post("/validate-otp", authHandler.ValidateOTP)

	// Protected routes (require JWT authentication)
	protected := v1.Group("/", middleware.JWTAuth())

	// User API (protected)
	userAPI := protected.Group("/users")
	userHandler := httpUser.NewUserHTTPHandler(userApp)
	userAPI.Get("/", userHandler.GetUsers)
	userAPI.Get("/:id", userHandler.GetUserByID)
	userAPI.Post("/", userHandler.CreateUser)
	userAPI.Put("/:id", userHandler.UpdateUser)
	userAPI.Delete("/:id", userHandler.DeleteUser)
	userAPI.Post("/change-password", userHandler.ChangePassword)
	userAPI.Post("/change-username", userHandler.ChangeUserName)
	userAPI.Post("/upload-avatar", userHandler.UploadAvatar)
	userAPI.Post("/change-phone", userHandler.ChangePhone)
	userAPI.Post("/change-email", userHandler.ChangeEmail)

	// Media API (protected)
	mediaAPI := protected.Group("/media")
	mediaHandler := httpMedia.NewMediaHttpHandler(mediaApp)
	mediaAPI.Get("/", mediaHandler.GetMediaList)
	mediaAPI.Get("/:id", mediaHandler.GetMediaId)
	mediaAPI.Post("/", mediaHandler.CreateMedia)

	// Setting API (protected)
	settingAPI := protected.Group("/settings")
	settingHandler := httpSetting.NewSettingHTTPHandler(settingApp)
	settingAPI.Get("/", settingHandler.GetAllSettings)
	settingAPI.Get("/:key", settingHandler.GetSettingByKey)
	settingAPI.Post("/", settingHandler.CreateSetting)
	settingAPI.Put("/:key", settingHandler.UpdateSetting)
	settingAPI.Delete("/:key", settingHandler.DeleteSetting)

	// Auth logout (protected)
	authApi.Post("/logout", middleware.JWTAuth(), authHandler.Logout)

	// Queue API (protected)
	queueAPI := protected.Group("/queues")
	queueHandler := httpQueue.NewQueueHTTPHandler(queueApp)
	queueAPI.Get("/", queueHandler.GetQueues)
	// queueAPI.Get("/:key", queueHandler.GetQueueByKey)

	// Tag API (protected)
	tagAPI := protected.Group("/tags")
	tagHandler := httTag.NewTagHttpHandler(tagApp)
	tagAPI.Post("/", tagHandler.CreateTag)
	tagAPI.Get("/", tagHandler.GetAllTags)
	tagAPI.Get("/:id", tagHandler.GetTagByID)
	tagAPI.Put("/:id", tagHandler.UpdateTag)
	tagAPI.Delete("/:id", tagHandler.DeleteTag)

	// Post API (protected)
	postAPI := protected.Group("/posts")
	postHandler := httpPost.NewPostHttpHandler(postApp)
	postAPI.Post("/", postHandler.CreatePost)
	postAPI.Get("/", postHandler.GetAllPosts)
	postAPI.Get("/:id", postHandler.GetPostByID)
	postAPI.Put("/:id", postHandler.UpdatePost)
	postAPI.Delete("/:id", postHandler.DeletePost)

	// Error Case Handler
	miscellaneousHandler := httpMiscellaneous.NewMiscellaneousHTTPHandler()
	r.All("*", miscellaneousHandler.NotFound)
}
