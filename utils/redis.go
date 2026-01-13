package utils

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func SetupRedis(ctx context.Context) {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 检查连接
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
}

func GetRedisClient() *redis.Client {
	return rdb
}
