package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func FetchPolicies(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, query string) map[string]*PolicySchema {
	data := make(map[string]*PolicySchema)

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching policies from the database")
	}

	defer rows.Close()

	for rows.Next() {
		var policy *PolicySchema
		if err := rows.Scan(&policy); err != nil {
			log.Error().Err(err).Msg("Error scanning policy from the database")
			continue
		}

		cacheKey := policy.Scope + ":" + policy.Identifier
		data[cacheKey] = policy
	}

	return data
}


func FetchPolicyByKey(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, query, cacheKey string) (*PolicySchema, bool) {
	var policy *PolicySchema
	err := db.QueryRow(ctx, query, cacheKey).Scan(&policy)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching policy by key from the database")
		return nil, false
	}

	return policy, true
}