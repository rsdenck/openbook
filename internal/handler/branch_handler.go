package handler

import (
	"openbook/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BranchHandler struct {
	uc *usecase.GitUseCase
}

func NewBranchHandler(uc *usecase.GitUseCase) *BranchHandler {
	return &BranchHandler{uc: uc}
}

func (h *BranchHandler) Create(c *fiber.Ctx) error {
	var req struct {
		SiteID     string `json:"site_id"`
		Name       string `json:"name"`
		FromCommit string `json:"from_commit"` // Optional
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

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

	var fromCommitID *uuid.UUID
	if req.FromCommit != "" {
		id, err := uuid.Parse(req.FromCommit)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid from_commit ID"})
		}
		fromCommitID = &id
	}

	branch, err := h.uc.CreateBranch(c.Context(), workspaceID, siteID, req.Name, fromCommitID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(branch)
}

func (h *BranchHandler) Get(c *fiber.Ctx) error {
	siteIDStr := c.Query("site_id")
	name := c.Query("name")

	if siteIDStr == "" || name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "site_id and name are required"})
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid site ID"})
	}

	branch, err := h.uc.GetBranch(c.Context(), siteID, name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branch not found"})
	}

	return c.JSON(branch)
}
