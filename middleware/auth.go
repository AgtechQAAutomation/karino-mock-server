package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shyamsundaar/karino-mock-server/initializers"
)

// ApiKeyAuth validates the X-API-KEY header
func ApiKeyAuth(c *fiber.Ctx) error {
	expectedKey := initializers.AppConfig.ApiKey
	clientKey := c.Get("APIKey")

	if clientKey == "" || clientKey != expectedKey {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid or missing API Key",
		})
	}
	return c.Next()
}

func JSONProviderMiddleware(c *fiber.Ctx) error {
	contentType := c.Get("Content-Type")

	// If header is missing, we allow it (as per your request)
	if contentType == "" {
		return c.Next()
	}

	// If header is present, it MUST contain application/json
	// We use strings.Contains because some clients send "application/json; charset=utf-8"
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"status":  "fail",
			"message": "Unsupported Media Type. If Content-Type is provided, it must be application/json",
		})
	}

	return c.Next()
}
