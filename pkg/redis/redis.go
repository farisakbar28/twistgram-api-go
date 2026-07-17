package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	Client = client
	return client
}

func GetClient() *redis.Client {
	return Client
}
