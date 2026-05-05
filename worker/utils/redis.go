package utils

import (
	"context"
	"encoding/json"
	"worker/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)

	//fail fast
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func PublishMessage(ctx context.Context, client *redis.Client, queryId string, msg WSMessage) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return client.Publish(ctx, queryId, string(bytes)).Err()
}
