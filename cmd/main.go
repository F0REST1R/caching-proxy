package main

import (
	"caching-proxy/internal/cache"
	"caching-proxy/internal/proxy"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := os.Getenv("PROXY_PORT")
	targetUrl, exist := os.LookupEnv("TARGET_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	clearCache := getEnvAsBool("CLEAR_CACHE", false)

	if !exist {
		log.Fatal("ERROR: URL environment variable is required")
	} 

	cacheS, err := cache.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}
	defer cacheS.Close()

	ctx := context.Background()
	if clearCache {
		if err := cacheS.Clear(ctx); err != nil{
			log.Printf("Warning: cache clear failed: %v", err)
		} else {
			log.Println("✓ Cache cleared successfully")
		}
	}

	handler := &proxy.HandlerProxy{
		TargetUrl: targetUrl,
		Cache: cacheS,
	}

	log.Printf("Starting caching proxy on port %s", port)
	log.Printf("Target server: %s", targetUrl)
	log.Printf("Redis cache: %s", redisAddr)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func getEnvAsBool(key string, defaultValue bool) bool{
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

//CLEAR_CACHE=true PORT=8080 TARGET_URL=https://jsonplaceholder.typicode.com docker-compose up