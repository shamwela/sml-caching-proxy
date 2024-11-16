package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sml-caching-proxy/helpers"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0, // default database
		Password: "",
	})
	fmt.Println("redisClient:", redisClient)

	shouldClearCache := flag.Bool("clear-cache", false, "Clear the cache.")
	port := flag.Int("port", 3000, "Port to listen on")
	origin := flag.String("origin", "", "Origin to proxy requests to")
	flag.Parse()

	if *shouldClearCache {
		helpers.ClearCache(redisClient)
		return
	}

	if *origin == "" {
		panic("--origin flag is required.")
	}

	isOriginValid := strings.HasPrefix(*origin, "http://") || strings.HasPrefix(*origin, "https://")

	if !isOriginValid {
		panic("\"origin\" must start with http:// or https://")
	}

	// If the origin ends with a slash, remove it.
	formattedOrigin := strings.TrimRight(*origin, "/")
	fmt.Println("formattedOrigin:", formattedOrigin)

	app := fiber.New()

	app.Use("/", func(context *fiber.Ctx) error {
		if context.Method() != "GET" {
			return context.Status(405).SendString("This caching proxy only supports GET requests.")
		}

		path := context.Path()
		fmt.Println("path:", path)

		fullUrl := formattedOrigin + path
		fmt.Println("finalUrl:", fullUrl)

		cache, cacheError := helpers.GetCache(redisClient, fullUrl)

		if cacheError != nil {
			return context.Status(500).SendString("Failed to get the cache.")
		}

		if cache != nil && cache != "" {
			context.Set("X-Cache", "HIT")
			context.Set("Content-Type", "application/json")
			return context.Status(200).JSON(cache)
		}

		// Cache miss.
		originResponse, originError := http.Get(fullUrl)

		if originError != nil {
			return context.Status(500).SendString(fmt.Sprintf("Failed to make the request to %s.", fullUrl))
		}

		defer originResponse.Body.Close()
		body, ioReadAllError := io.ReadAll(originResponse.Body)

		if ioReadAllError != nil {
			return context.Status(500).SendString("Failed to read the origin's response body.")
		}

		if body == nil {
			return context.Status(500).SendString("Origin returned an empty response body.")
		}

		// Verify it's a valid JSON before caching.
		var jsonTest json.RawMessage
		unmarshalError := json.Unmarshal(body, &jsonTest)

		if unmarshalError != nil {
			return context.Status(500).SendString("Origin returned an invalid JSON.")
		}

		const expirationTime = 10 * time.Minute
		// Use the full URL as the cache key.
		redisSetError := redisClient.Set(fullUrl, body, expirationTime).Err()

		if redisSetError != nil {
			return context.Status(500).SendString("Failed to store the origin's response body in Redis.")
		}

		fmt.Println("Cache miss.")
		context.Set("X-Cache", "MISS")
		context.Set("Content-Type", "application/json")
		return context.Status(originResponse.StatusCode).Send(body)
	})

	app.Listen(fmt.Sprintf(":%d", *port))
}
