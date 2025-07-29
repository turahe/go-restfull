package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterPostRoutes registers all post-related routes
func RegisterPostRoutes(protected fiber.Router, container *container.Container) {
	postController := container.GetPostController()

	// Post routes (protected)
	posts := protected.Group("/posts")
	posts.Post("/", postController.CreatePost)
	posts.Get("/", postController.GetPosts)
	posts.Get("/:id", postController.GetPostByID)
	posts.Get("/slug/:slug", postController.GetPostBySlug)
	posts.Get("/author/:authorID", postController.GetPostsByAuthor)
	posts.Put("/:id", postController.UpdatePost)
	posts.Delete("/:id", postController.DeletePost)
	posts.Put("/:id/publish", postController.PublishPost)
	posts.Put("/:id/unpublish", postController.UnpublishPost)
}
