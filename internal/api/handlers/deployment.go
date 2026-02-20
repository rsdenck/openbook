package handlers

import (
	"openbook/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DeploymentHandler struct {
	publishingUseCase *usecase.PublishingUseCase
}

func NewDeploymentHandler(publishingUseCase *usecase.PublishingUseCase) *DeploymentHandler {
	return &DeploymentHandler{publishingUseCase: publishingUseCase}
}

// PublishSpace godoc
// @Summary Trigger a publish for a space
// @Description Creates a new deployment and starts the build process
// @Tags deployments
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param space_id path string true "Space ID"
// @Success 202 {object} domain.Deployment
// @Router /workspaces/{workspace_id}/spaces/{space_id}/publish [post]
func (h *DeploymentHandler) PublishSpace(c *fiber.Ctx) error {
	workspaceID, err := uuid.Parse(c.Params("workspace_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid workspace_id"})
	}

	spaceID, err := uuid.Parse(c.Params("space_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid space_id"})
	}

	// In a real app, get user ID from context (JWT middleware)
	// userID := c.Locals("user_id").(uuid.UUID)
	userID := uuid.Nil // Mock user for now

	deployment, err := h.publishingUseCase.TriggerPublish(c.Context(), workspaceID, spaceID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(deployment)
}
