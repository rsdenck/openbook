package routes

import (
	"openbook/internal/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, deploymentHandler *handlers.DeploymentHandler) {
	api := app.Group("/api/v1")

	// Workspaces
	workspaces := api.Group("/workspaces")
	
	// Spaces within a workspace
	spaces := workspaces.Group("/:workspace_id/spaces")
	
	// Publish endpoint
	spaces.Post("/:space_id/publish", deploymentHandler.PublishSpace)

	// Add other routes for Pages, Collections, etc.
}
