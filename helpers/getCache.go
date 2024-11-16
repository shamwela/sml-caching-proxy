package helpers

import (
	"encoding/json"
	"errors"

	"github.com/go-redis/redis"
)

func GetCache(redisClient *redis.Client, fullUrl string) (cacheData interface{}, error error) {
	cacheString := redisClient.Get(fullUrl).Val()

	if cacheString == "" {
		return nil, nil
	}

	cacheByteSlice := []byte(cacheString)
	var cache interface{}
	unmarshalError := json.Unmarshal(cacheByteSlice, &cache)

	if unmarshalError != nil {
		return nil, errors.New("failed to unmarshal the cache")
	}

	return cache, nil
}
