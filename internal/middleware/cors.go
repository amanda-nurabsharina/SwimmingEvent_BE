package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupCORS(allowOrigins string) fiber.Handler {
	origins := strings.Split(allowOrigins, ",")
	originMap := make(map[string]bool)
	for _, o := range origins {
		trimmed := strings.ToLower(strings.TrimSpace(o))
		if trimmed != "" {
			originMap[trimmed] = true
		}
	}

	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == "" {
				return true
			}
			lower := strings.ToLower(origin)

			// If configured to allow all or "*"
			if originMap["*"] || allowOrigins == "*" {
				return true
			}

			// Explicit match in ALLOWED_ORIGINS
			if originMap[lower] {
				return true
			}

			// Localhost / 127.0.0.1 for local development
			if strings.HasPrefix(lower, "http://localhost:") || strings.HasPrefix(lower, "http://127.0.0.1:") || lower == "http://localhost" || lower == "http://127.0.0.1" {
				return true
			}

			// Allow all subdomains of fourplusone.my.id or the domain itself
			if strings.HasSuffix(lower, ".fourplusone.my.id") || strings.Contains(lower, "fourplusone.my.id") {
				return true
			}

			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Requested-With",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

