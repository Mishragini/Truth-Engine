package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisUrl   string
	DBURL     string
	GroqApiKey string
	JinaApiKey string
}

func Load() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		return nil, err
	}

	env := Config{}

	redisUrl := os.Getenv("REDIS_URL")
	dbUrl := os.Getenv("DATABASE_URL")
	groqApiKey := os.Getenv("GROQ_API_KEY")
	jinaApiKey := os.Getenv("JINA_API_KEY")
	env.RedisUrl = redisUrl
	env.DBURL= dbUrl
	env.GroqApiKey = groqApiKey
	env.JinaApiKey = jinaApiKey

	return &env, nil
}
