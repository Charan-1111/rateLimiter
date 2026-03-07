package handlers

import (
	"goapp/constants"
	"goapp/logic"

	"github.com/gofiber/fiber/v2"
)

func (cfg *ConfigHandler) GetLimiter(c *fiber.Ctx) error {
	queries := c.Queries()
	limiterType := queries[constants.KeyRateLimitType]
	algorithm := queries[constants.KeyAlgo]

	// validate the limiter type and algorithm
	if limiterType == "" || algorithm == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required query parameters: type and algo",
		})
	}

	logic.GetLimiter(cfg.factory, limiterType, algorithm)
	return nil
}
