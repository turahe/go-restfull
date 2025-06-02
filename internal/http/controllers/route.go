package httpapi

import (
	"github.com/gofiber/fiber/v2"
	"webapi/internal/app/media"
	"webapi/internal/app/queue"
	"webapi/internal/app/user"
	"webapi/internal/repository"

	httpAuth "webapi/internal/http/controllers/auth"
	httpHealthz "webapi/internal/http/controllers/healthz"
	httpMedia "webapi/internal/http/controllers/media"
	httpMiscellaneous "webapi/internal/http/controllers/miscellaneous"
	httpQueue "webapi/internal/http/controllers/queue"
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

	// User API
	userAPI := v1.Group("/users")
	userApp := user.NewUserApp(repo)
	userHandler := httpUser.NewUserHTTPHandler(userApp)
	userAPI.Get("/", userHandler.GetUsers)
	userAPI.Get("/:id", userHandler.GetUserByID)
	userAPI.Post("/", userHandler.CreateUser)
	userAPI.Put("/:id", userHandler.UpdateUser)
	userAPI.Delete("/:id", userHandler.DeleteUser)
	userAPI.Post("/change-password", userHandler.ChangePassword)
	userAPI.Post("/change-username", userHandler.ChangeUserName)
	userAPI.Post("/change-phone", userHandler.ChangePhone)
	userAPI.Post("/change-email", userHandler.ChangeEmail)

	// Media API
	mediaAPI := v1.Group("/media")
	mediaApp := media.NewMediaApp(repo)
	mediaHandler := httpMedia.NewMediaHttpHandler(mediaApp)
	mediaAPI.Get("/", mediaHandler.GetMediaList)
	mediaAPI.Get("/:id", mediaHandler.GetMediaId)
	mediaAPI.Post("/", mediaHandler.CreateMedia)

	// auth
	authApi := v1.Group("/auth")
	registerHandler := httpAuth.NewRegisterHTTPHandler(userApp)
	authApi.Post("/register", registerHandler.Register)
	loginHandler := httpAuth.NewLoginHTTPHandler(userApp)
	authApi.Post("/login", loginHandler.Login)

	// Queue API
	queueAPI := v1.Group("/queues")
	queueApp := queue.NewQueueApp(repo)
	queueHandler := httpQueue.NewQueueHTTPHandler(queueApp)
	queueAPI.Get("/", queueHandler.GetQueues)
	// queueAPI.Get("/:key", queueHandler.GetQueueByKey)

	// Error Case Handler
	miscellaneousHandler := httpMiscellaneous.NewMiscellaneousHTTPHandler()
	r.All("*", miscellaneousHandler.NotFound)
}
