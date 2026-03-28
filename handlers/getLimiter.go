package handlers

import (
	"goapp/constants"
	"goapp/logic"
	"goapp/logger"
	"strconv"

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

	reqLog := logger.GetRequestLogger(c, cfg.log)
	allowed, err := logic.GetLimiter(cfg.ctx, cfg.db, cfg.rdb, cfg.config, reqLog, cfg.factory, cfg.cache, cfg.cb, scope, identifier, rateLimitType)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error while evaluating rate limits",
		})
	}

	// adding to headers
	c.Set("X-RateLimit-Limit", strconv.FormatInt(allowed.TotalTokens, 10))
	c.Set("X-RateLimit-Remaining", strconv.FormatInt(allowed.RemainingTokens, 10))
	c.Set("X-RateLimit-Retry-After", strconv.FormatInt(allowed.RetryAfter, 10))

	if !allowed.Allowed {
		return c.Status(fiber.StatusTooManyRequests).JSON(allowed)
	}

	return c.Status(fiber.StatusOK).JSON(allowed)
}
