package middleware

import (
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
