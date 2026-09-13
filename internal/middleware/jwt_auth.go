package middleware

import (
	"strings"

	"serve-swimming-be/pkg/response"
	"serve-swimming-be/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func RequireJWT(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Authorization token required", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid authorization header format", nil)
		}

		claims, err := security.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or expired JWT token", err.Error())
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
