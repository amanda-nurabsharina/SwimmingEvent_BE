package middleware

import (
	"serve-swimming-be/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func RequireAPIKey(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != secretKey {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		return c.Next()
	}
}
