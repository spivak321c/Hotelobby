package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Context key types for Fiber Locals — used by AuthMiddleware and handlers.
type contextKey string

const (
	KeyUserID = contextKey("user_id")
	KeyRole   = contextKey("role")
)

// TokenValidator is satisfied by *services.AuthService.
type TokenValidator interface {
	ValidateToken(tokenStr string) (userID, role string, err error)
}

// AuthMiddleware extracts and validates the Bearer JWT from Authorization header.
// On success it stores user_id and role in Fiber Locals.
func AuthMiddleware(validator TokenValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrResp("missing_token", "missing authorization header"))
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrResp("invalid_token_format", "invalid authorization format"))
		}

		userID, role, err := validator.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrResp("invalid_token", "invalid or expired token"))
		}

		c.Locals(KeyUserID, userID)
		c.Locals(KeyRole, role)
		return c.Next()
	}
}

// RoleMiddleware allows only the specified roles through.
// Must be used after AuthMiddleware.
func RoleMiddleware(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(KeyRole).(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(ErrResp("forbidden", "forbidden"))
		}
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(ErrResp("insufficient_permissions", "insufficient permissions"))
	}
}
