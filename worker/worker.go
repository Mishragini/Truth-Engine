package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"worker/db"
	"worker/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
)

func worker(ctx context.Context, result string, redisClient *redis.Client, pubsubClient *redis.Client, dbPool *pgxpool.Pool, httpClient *http.Client, openAiClient *openai.Client, jinaApiKey string) error {
	var query utils.Query
	err := json.Unmarshal([]byte(result), &query)
	if err != nil {
		return err
	}

	existingRes, err := db.GetExistingResponse(ctx, dbPool, query.SearchQuery)
	if err != nil {
		return err
	}

	if existingRes != nil {
		msg := utils.WSMessage{
			Type:    "full_response",
			Payload: existingRes.Response,
		}
		err = utils.PublishMessage(ctx, pubsubClient, query.QueryId, msg)
		if err != nil {
			return err
		}
		return nil
	}

	embedding, err := utils.GetEmbedding(ctx, query.SearchQuery, httpClient, jinaApiKey)

	if err != nil {
		return err
	}

	data, err := utils.SemanticSearch(ctx, embedding, dbPool)

	if err != nil {
		return err
	}

	if len(data) <= 0 {
		keywords, err := utils.ExtractKeywords(ctx, openAiClient, query.SearchQuery)
		if err != nil {
			return err
		}
		hackernews, err := utils.FetchHackerNews(ctx, keywords, httpClient)
		if err != nil {
			return err
		}
		sr := utils.SearchResult{
			Query:     query.SearchQuery,
			Response:  "",
			Resources: hackernews,
		}
		data = append(data, sr)
	}

	var resources []*utils.Resource
	for _, result := range data {
		resources = append(resources, result.Resources...)
	}
	stream, err := utils.PromptLlm(ctx, openAiClient, resources, query.SearchQuery)
	var b strings.Builder
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(chunk.Choices) == 0 {
			return fmt.Errorf("no choices returned from AI")
		}

		token := chunk.Choices[0].Delta.Content

		fmt.Fprintf(&b, "%v", token)

		msg := utils.WSMessage{
			Type:    "chunk",
			Payload: token,
		}
		err = utils.PublishMessage(ctx, pubsubClient, query.QueryId, msg)
		if err != nil {
			return err
		}
	}

	aiResponse := b.String()

	saveResponse := utils.SaveResponseDb{
		Query:     query.SearchQuery,
		Embedding: *embedding,
		Response:  aiResponse,
		Resources: resources,
	}
	err = db.SaveResponse(ctx, dbPool, saveResponse)
	if err != nil {
		return err
	}
	cacheValue, err := json.Marshal(aiResponse)
	if err != nil {
		return err
	}
	if err = redisClient.Set(ctx, query.SearchQuery, cacheValue, time.Hour*24).Err(); err != nil {
		return err
	}

	return nil
}
