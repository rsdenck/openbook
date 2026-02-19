package handler

import (
	"openbook/internal/domain"
	"openbook/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DeploymentHandler struct {
	uc *usecase.DeploymentUseCase
}

func NewDeploymentHandler(uc *usecase.DeploymentUseCase) *DeploymentHandler {
	return &DeploymentHandler{uc: uc}
}

func (h *DeploymentHandler) Create(c *fiber.Ctx) error {
	var req struct {
		SiteID        string `json:"site_id"`
		EnvironmentID string `json:"environment_id"`
		CommitHash    string `json:"commit_hash"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Extract WorkspaceID from context (middleware)
	workspaceIDStr, ok := c.Locals("workspace_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid workspace ID"})
	}

	siteID, err := uuid.Parse(req.SiteID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid site ID"})
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid environment ID"})
	}

	deployment := &domain.Deployment{
		WorkspaceID:   workspaceID,
		SiteID:        siteID,
		EnvironmentID: envID,
		CommitHash:    req.CommitHash,
		// TriggeredBy: userID from context (TODO)
	}

	if err := h.uc.Create(c.Context(), deployment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(deployment)
}

func (h *DeploymentHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	d, err := h.uc.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Deployment not found"})
	}

	return c.JSON(d)
}
