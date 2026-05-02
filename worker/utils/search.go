package utils

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type SearchResult struct {
	Query     string
	Response  string
	Resources []*Resource
}

func SemanticSearch(ctx context.Context, embedding *pgvector.Vector, dbPool *pgxpool.Pool) ([]SearchResult, error) {
	sqlQuery := `
				SELECT 
				sr.query, sr.response,
				r.id,r.title,
				s.content
				FROM search_responses sr
				LEFT JOIN resources r ON r.search_response_id = sr.id
				LEFT JOIN sources s ON s.resource_id = r.id
				WHERE valid_till > now() 
				AND 1 - (query_embedding <=> $1) >= 0.75
				ORDER BY query_embedding <=> $1
	 			LIMIT 5
				`

	rows, err := dbPool.Query(ctx, sqlQuery, *embedding)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type resourceKey struct{ query, resourceID string }
	resultMap := make(map[string]*SearchResult)
	resourceMap := make(map[resourceKey]*Resource)

	var order []string

	for rows.Next() {
		var (
			query      string
			response   string
			resourceID *string
			title      *string
			content    *[]string
		)
		if err := rows.Scan(&query, &response, &resourceID, &title, &content); err != nil {
			return nil, err
		}

		if _, exists := resultMap[query]; !exists {
			resultMap[query] = &SearchResult{
				Query:    query,
				Response: response,
			}
			order = append(order, query)
		}
		sr := resultMap[query]
		if resourceID == nil {
			continue
		}

		rKey := resourceKey{query, *resourceID}
		if _, exists := resourceMap[rKey]; !exists {
			resource := &Resource{Title: *title}
			resourceMap[rKey] = resource
			sr.Resources = append(sr.Resources, resource)
		}

		res := resourceMap[rKey]

		if content != nil {
			res.Content = *content
		}
	}

	results := make([]SearchResult, 0, len(order))

	for _, q := range order {
		results = append(results, *resultMap[q])
	}
	return results, nil
}
