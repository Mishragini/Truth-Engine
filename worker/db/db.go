package db

import (
	"context"
	"fmt"
	"time"
	"worker/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

func NewDbPool(ctx context.Context, dbUrl string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnIdleTime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	return pool, nil
}

type SearchResponse struct {
	Response  string
	Embedding pgvector.Vector
	ValidTill time.Time
}

func GetExistingResponse(ctx context.Context, dbPool *pgxpool.Pool, query string) (*SearchResponse, error) {
	var existingRes SearchResponse
	sqlQuery := `SELECT response,query_embedding,valid_till FROM search_responses WHERE query = $1 AND valid_till > now()`
	err := dbPool.QueryRow(ctx, sqlQuery, query).Scan(&existingRes.Response, &existingRes.Embedding, &existingRes.ValidTill)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &existingRes, nil
}

func SaveResponse(ctx context.Context, dbPool *pgxpool.Pool, searchResponse utils.SaveResponseDb) error {
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var searchResponseID string
	err = tx.QueryRow(ctx, `
		INSERT INTO search_responses(query,query_embedding,response,valid_till)
		VALUES ($1,$2,$3,now() + interval '24 hours')
		ON CONFLICT (query) DO UPDATE
			SET query_embedding = EXCLUDED.query_embedding,
				response = EXCLUDED.response,
				valid_till = EXCLUDED.valid_till
		RETURNING id`,
		searchResponse.Query,
		searchResponse.Embedding,
		searchResponse.Response,
	).Scan(&searchResponseID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        DELETE FROM resources WHERE search_response_id = $1
    `, searchResponseID)
	if err != nil {
		return fmt.Errorf("clear stale resources: %w", err)
	}

	for _, resource := range searchResponse.Resources {
		var resourceID string
		err = tx.QueryRow(ctx, `
		INSERT INTO resources (title,search_response_id)
		VALUES($1,$2) RETURNING id`,
			resource.Title,
			searchResponseID,
		).Scan(&resourceID)

		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO sources (resource_id,content)
			VALUES($1,$2)
			`, resourceID, resource.Content)

		if err != nil {
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
