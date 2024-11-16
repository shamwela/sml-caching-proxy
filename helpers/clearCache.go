package helpers

import (
	"fmt"

	"github.com/go-redis/redis"
)

func ClearCache(redisClient *redis.Client) {
	redisFlushAllError := redisClient.FlushAll().Err()

	if redisFlushAllError != nil {
		panic("Failed to clear the cache.")
	}

	fmt.Println("Successfully cleared the cache.")
}
