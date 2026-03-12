package handlers

import (
	"goapp/constants"
	"goapp/logic"

	"github.com/gofiber/fiber/v2"
)

func (cfg *ConfigHandler) GetLimiter(c *fiber.Ctx) error {
	queries := c.Queries()
	scope := queries[constants.KeyScope]
	identifier := queries[constants.KeyIdentifier]
	rateLimitType := queries[constants.KeyRateLimitType]
	
	// validate the limiter type and algorithm
	if scope == "" || identifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required query parameters: scope and identifier",
		})
	}

	logic.GetLimiter(c.Context(), cfg.db, cfg.rdb, cfg.config, cfg.log, cfg.factory, cfg.cache, scope, identifier, rateLimitType)
	return nil
}
