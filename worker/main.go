package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"worker/config"
	"worker/db"
	"worker/utils"

	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()

	if err != nil {
		log.Printf("failed to load config: %v", err)
		return
	}

	redisClient, err := utils.NewRedisClient(ctx, cfg)
	if err != nil {
		log.Printf("failed to establish queue client: %v", err)
		return
	}
	defer redisClient.Close()

	dbUrl := cfg.DBURL
	dbPool, err := db.NewDbPool(ctx, dbUrl)
	if err != nil {
		log.Printf("failed to connect to db: %v", err)
		return
	}
	defer dbPool.Close()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	pubsubClient, err := utils.NewRedisClient(ctx, cfg)
	if err != nil {
		log.Printf("failed to establish pubsub client %v", err)
		return
	}
	defer pubsubClient.Close()

	openaiConfig := openai.DefaultConfig(cfg.GroqApiKey)
	openaiConfig.BaseURL = "https://api.groq.com/openai/v1"
	openAiClient := openai.NewClientWithConfig(openaiConfig)

	const numWorkers = 5
	jobs := make(chan string, 10)

	for range numWorkers {
		go func() {
			for job := range jobs {
				workerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				err := worker(workerCtx, job, redisClient,pubsubClient, dbPool, httpClient, openAiClient, cfg.JinaApiKey)
				cancel()
				if err != nil {
					log.Printf("worker failed: %v", err)
				}
			}
		}()
	}

	for {
		result, err := redisClient.BRPop(ctx, 0, "search_query").Result()
		if err != nil {
			log.Printf("failed to pop from queue: %v", err)
			continue
		}
		jobs <- result[1]
	}

}
