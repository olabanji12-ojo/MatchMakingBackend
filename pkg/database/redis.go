package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(addr, password string) *redis.Client {
	var opt *redis.Options
	var err error

	// Check if the address is a full URL (common for Upstash/Managed Redis)
	if len(addr) > 8 && (addr[:8] == "redis://" || addr[:9] == "rediss://") {
		opt, err = redis.ParseURL(addr)
		if err != nil {
			log.Fatalf("Failed to parse Redis URL: %v", err)
		}
	} else {
		opt = &redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}
	}

	client := redis.NewClient(opt)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis established")
	RedisClient = client
	return client
}
