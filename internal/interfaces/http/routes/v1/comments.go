package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterCommentRoutes registers all comment-related routes
func RegisterCommentRoutes(protected fiber.Router, container *container.Container) {
	commentController := container.GetCommentController()

	// Comment routes (protected)
	comments := protected.Group("/comments")
	comments.Get("/", commentController.GetComments)
	comments.Get("/:id", commentController.GetCommentByID)
	comments.Post("/", commentController.CreateComment)
	comments.Put("/:id", commentController.UpdateComment)
	comments.Delete("/:id", commentController.DeleteComment)
	comments.Put("/:id/approve", commentController.ApproveComment)
	comments.Put("/:id/reject", commentController.RejectComment)
}
