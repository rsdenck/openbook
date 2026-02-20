package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantContextKey is the key used to store/retrieve the workspace ID from context
type TenantContextKey string

const (
	WorkspaceIDKey TenantContextKey = "workspace_id"
	UserIDKey      TenantContextKey = "user_id"
	RoleKey        TenantContextKey = "role"
)

// TenantMiddleware extracts the workspace_id from the request (header, query, or path)
// and injects it into the context. It validates that the tenant exists.
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var workspaceIDStr string

		// 1. Try to get from Path Parameter (highest priority)
		workspaceIDStr = c.Params("workspace_id")

		// 2. Try to get from Header (X-Workspace-ID)
		if workspaceIDStr == "" {
			workspaceIDStr = c.Get("X-Workspace-ID")
		}

		// 3. Try to get from Query
		if workspaceIDStr == "" {
			workspaceIDStr = c.Query("workspace_id")
		}

		if workspaceIDStr == "" {
			// For some routes (like creating a workspace or login), tenant might not be needed yet.
			// But for protected resources, it's mandatory.
			// We can decide to skip or fail. For now, let's continue but check later in handlers.
			return c.Next()
		}

		id, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid workspace_id format",
			})
		}

		// Inject into Fiber Locals (fast access)
		c.Locals(string(WorkspaceIDKey), id)
		
		// Inject into Standard Context (for repository layer)
		ctx := context.WithValue(c.Context(), WorkspaceIDKey, id)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// RequireTenant ensures that a workspace ID is present in the context
func RequireTenant(c *fiber.Ctx) error {
	id := c.Locals(string(WorkspaceIDKey))
	if id == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Workspace ID is required for this operation",
		})
	}
	return c.Next()
}

// GetWorkspaceID retrieves the workspace ID from the context
func GetWorkspaceID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(WorkspaceIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("workspace_id not found in context")
	}
	return id, nil
}
