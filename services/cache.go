package services

import (
	"context"
	"goapp/constants"

	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type PolicySchema struct {
	Scope      string `json:"scope"`
	Identifier string `json:"identifier"`
	Limit      int    `json:"limit"`
	Window     string `json:"window"`
	Burst      int    `json:"burst"`
	Algorithm  string `json:"algorithm"`
}

type Cache struct {
	data *ristretto.Cache
}

func NewCache() *Cache {
	c, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,
		MaxCost:     1 << 28,
		BufferItems: 64,
	})
	return &Cache{
		data: c,
	}
}

func (c *Cache) LoadCache(ctx context.Context, log zerolog.Logger, db *pgxpool.Pool, query string) {
	policies := FetchPolicies(ctx, db, log, query)

	for policyKey, policy := range policies {
		c.data.SetWithTTL(policyKey, policy, 1, constants.PolicyCacheDuration)
	}
}

func (c *Cache) GetPolicy(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, scope, identifier, query string) (*PolicySchema, bool) {
	cacheKey := scope + ":" + identifier

	policy, exists := c.data.Get(cacheKey)

	if !exists {
		// Fetch from the database
		policy, exists = FetchPolicyByKey(ctx, db, log, query, cacheKey)
		// store in the cache
		if exists {
			// Double-check locking to avoid overwriting if another goroutine fetched it concurrently
			if existingPolicy, ok := c.data.Get(cacheKey); ok {
				policy = existingPolicy
			} else {
				c.data.SetWithTTL(cacheKey, policy, 1, constants.PolicyCacheDuration)
			}
		}
	}
	return policy.(*PolicySchema), exists
}
