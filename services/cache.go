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

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching policies from the database")
	}

	defer rows.Close()

	for rows.Next() {
		var policy PolicySchema
		if err := rows.Scan(&policy); err != nil {
			log.Error().Err(err).Msg("Error scanning policy from the database")
			continue
		}

		cacheKey := policy.Scope + ":" + policy.Identifier
		c.data[cacheKey] = &policy
	}
}

func (c *Cache) GetPolicy(scope, identifier string) (*PolicySchema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheKey := scope + ":" + identifier
	policy, exists := c.data[cacheKey]
	return policy, exists
}
