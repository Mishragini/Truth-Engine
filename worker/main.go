package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"worker/config"
	"worker/db"
	"worker/utils"

	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//channel that will receive the signal
	//buffer 1 to handle that the async signal sending
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

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

	//wait group for worker goroutines to wait for them to finish before returning
	var wg sync.WaitGroup

	const numWorkers = 5
	jobs := make(chan string, 10)

	for range numWorkers {
		//add to the waitgroup
		wg.Add(1)
		go func() {
			//decrement from the waitgroup when the worker is done
			defer wg.Done()
			for job := range jobs {
				workerCtx, workerCancel := context.WithTimeout(ctx, 30*time.Second)
				err := worker(workerCtx, job, redisClient, pubsubClient, dbPool, httpClient, openAiClient, cfg.JinaApiKey)
				workerCancel()
				if err != nil {
					log.Printf("worker failed: %v", err)
				}
			}
		}()
	}

	brpopCtx, brpopCancel := context.WithCancel(context.Background())

	//separate go routine to handle the poppin out from the queue if we kept the infinit loop in the main go routine the code below it is unreachable
	go func() {
		for {
			result, err := redisClient.BRPop(brpopCtx, 0, "search_query").Result()
			if err != nil {
				if brpopCtx.Err() != nil {
					log.Println("intake loop shutting down")
					return
				}
				log.Printf("failed to pop from queue: %v", err)
				continue
			}
			jobs <- result[1]
		}
	}()

	// block main goroutine until shutdown signal is received
	sig := <-quit
	log.Printf("received signal: %v — shutting down", sig)

	//it's import to cancel the intake first before closing the channel
	//if not sending to closed channel results in panic
	brpopCancel()
	//signal worker goroutines there will no longer sends to jobs channel
	close(jobs)

	log.Println("waiting for workers to finish...")

	//wait for the worker go routine waitgroup
	wg.Wait()

	log.Println("shutdown complete")

}
