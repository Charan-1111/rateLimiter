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

	allowed, err := logic.GetLimiter(cfg.ctx, cfg.db, cfg.rdb, cfg.config, cfg.log, cfg.factory, cfg.cache, cfg.cb, scope, identifier, rateLimitType)
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error while evaluating rate limits",
		})
	}

	if !allowed.Allowed {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Too Many Requests",
			"retryAfter": allowed.RetryAfter,
			"currentTokens": allowed.CurrentTokens,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Request allowed",
		"retryAfter": allowed.RetryAfter,
		"currentTokens": allowed.CurrentTokens,
	})
}
