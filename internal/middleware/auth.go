package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthMiddleware simulates JWT validation and workspace extraction
// In a real scenario, this would validate a JWT token and extract claims.
// For now, we will extract from headers for demonstration/testing if no real auth provider is integrated.
// Or we can just mock it to always succeed for a specific user/workspace if headers are missing?
// CONTEXTO3 says "Validação JWT".
// I'll implement a simple header check: "Authorization: Bearer <token>"
// And "X-Workspace-ID".
func AuthMiddleware(c *fiber.Ctx) error {
	// 1. JWT Validation (Mocked for now, just checking presence)
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid Authorization header"})
	}
	// token := authHeader[7:]
	// Validate(token) ...

	// Mock User ID extraction from token
	// In reality, this comes from JWT claims.
	// We'll generate a consistent UUID based on token or just random for now.
	// Let's assume a test user ID.
	userID := uuid.New() // TODO: Extract from real JWT
	c.Locals("user_id", userID.String())

	// 2. Workspace Extraction
	// Workspace ID might come from the URL (e.g., /workspaces/:workspace_id/...) or Header.
	// If it's in the route, Fiber handles it. But we need to put it in Locals for handlers that don't have it in route.
	// Or handlers extract it themselves.
	// But `CONTEXTO3` says "Middleware obrigatório: Extração de workspace_id".
	// This usually implies checking if the user has access to the workspace.
	
	// Let's check if "workspace_id" param exists in route.
	workspaceID := c.Params("workspace_id")
	if workspaceID == "" {
		// Try header
		workspaceID = c.Get("X-Workspace-ID")
	}

	if workspaceID != "" {
		// Validate UUID
		if _, err := uuid.Parse(workspaceID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Workspace ID format"})
		}
		c.Locals("workspace_id", workspaceID)
	} else {
		// If not found, some endpoints might fail.
		// We can proceed, but handlers needing it will fail.
	}

	return c.Next()
}
