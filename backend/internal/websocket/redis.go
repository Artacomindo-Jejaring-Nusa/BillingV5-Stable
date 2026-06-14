package websocket

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisPubSub *redis.PubSub
var redisCtx = context.Background()

const redisChannel = "ws:notifications"

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("[WebSocket Redis] Failed to parse Redis URL: %v (pub/sub disabled)", err)
		return
	}

	redisClient = redis.NewClient(opts)

	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		log.Printf("[WebSocket Redis] Failed to connect: %v (pub/sub disabled)", err)
		redisClient = nil
		return
	}

	redisPubSub = redisClient.Subscribe(redisCtx, redisChannel)

	go listenRedisMessages()

	log.Println("[WebSocket Redis] Connected and subscribed to notification channel")
}

func PublishToRedis(payload map[string]interface{}) {
	if redisClient == nil {
		return
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WebSocket Redis] Failed to marshal payload: %v", err)
		return
	}
	if err := redisClient.Publish(redisCtx, redisChannel, bytes).Err(); err != nil {
		log.Printf("[WebSocket Redis] Failed to publish: %v", err)
	}
}

func listenRedisMessages() {
	if redisPubSub == nil {
		return
	}
	ch := redisPubSub.Channel()
	for msg := range ch {
		if GlobalHub == nil {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Printf("[WebSocket Redis] Failed to unmarshal message: %v", err)
			continue
		}
		if source, ok := payload["_source"].(string); ok && source == instanceID {
			continue
		}
		bytes, err := json.Marshal(payload)
		if err != nil {
			continue
		}
		GlobalHub.Broadcast <- bytes
	}
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func InvalidateDashboardCache(ctx context.Context) {
	if redisClient == nil {
		return
	}
	var cursor uint64
	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, "dashboard:cache:*", 100).Result()
		if err != nil {
			log.Printf("[Redis Cache] Error scanning keys for invalidation: %v", err)
			break
		}
		if len(keys) > 0 {
			redisClient.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	log.Println("[Redis Cache] Invalidated dashboard cache keys successfully")
}

var instanceID string

func init() {
	instanceID = os.Getenv("HOSTNAME")
	if instanceID == "" {
		instanceID = "unknown"
	}
	instanceID = instanceID + "_" + time.Now().Format("150405.000")
}
