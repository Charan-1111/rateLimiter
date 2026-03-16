package services

import (
	"context"
	"sync"

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
	mu   sync.RWMutex
	data map[string]*PolicySchema
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*PolicySchema),
	}
}

func (c *Cache) LoadCache(ctx context.Context, log zerolog.Logger, db *pgxpool.Pool, query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = FetchPolicies(ctx, db, log, query)
}

func (c *Cache) GetPolicy(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, scope, identifier, query string) (*PolicySchema, bool) {
	cacheKey := scope + ":" + identifier
	
	c.mu.RLock()
	policy, exists := c.data[cacheKey]
	c.mu.RUnlock()
	
	if !exists {
		// Fetch from the database
		policy, exists = FetchPolicyByKey(ctx, db, log, query, cacheKey)
		// store in the cache
		if exists {
			c.mu.Lock()
			// Double-check locking to avoid overwriting if another goroutine fetched it concurrently
			if existingPolicy, ok := c.data[cacheKey]; ok {
				policy = existingPolicy
			} else {
				c.data[cacheKey] = policy
			}
			c.mu.Unlock()
		}
	}
	return policy, exists
}
