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

const (
	redisChannel     = "ws:notifications"
	redisChatChannel = "ws:chat"
)

type RedisChatPayload struct {
	Source         string          `json:"_source"`
	RoomID         uint64          `json:"room_id"`
	Event          string          `json:"event"`
	Data           json.RawMessage `json:"data"`
	ExceptClientID string          `json:"except_client_id,omitempty"`
}

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

	redisPubSub = redisClient.Subscribe(redisCtx, redisChannel, redisChatChannel)

	go listenRedisMessages()

	log.Println("[WebSocket Redis] Connected and subscribed to notifications and chat channels")
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

// PublishChatToRedis menyinkronkan event chat ke seluruh instance backend server
func PublishChatToRedis(roomID uint64, event string, data interface{}, exceptClientID ...string) {
	if redisClient == nil {
		return
	}
	var excID string
	if len(exceptClientID) > 0 {
		excID = exceptClientID[0]
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("[WebSocket Redis Chat] Failed to marshal data: %v", err)
		return
	}

	payload := RedisChatPayload{
		Source:         instanceID,
		RoomID:         roomID,
		Event:          event,
		Data:           dataBytes,
		ExceptClientID: excID,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WebSocket Redis Chat] Failed to marshal payload: %v", err)
		return
	}

	if err := redisClient.Publish(redisCtx, redisChatChannel, bytes).Err(); err != nil {
		log.Printf("[WebSocket Redis Chat] Failed to publish chat event: %v", err)
	}
}

func listenRedisMessages() {
	if redisPubSub == nil {
		return
	}
	ch := redisPubSub.Channel()
	for msg := range ch {
		switch msg.Channel {
		case redisChannel:
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

		case redisChatChannel:
			if GlobalChatHub == nil {
				continue
			}
			var chatPayload RedisChatPayload
			if err := json.Unmarshal([]byte(msg.Payload), &chatPayload); err != nil {
				log.Printf("[WebSocket Redis Chat] Failed to unmarshal chat payload: %v", err)
				continue
			}
			if chatPayload.Source == instanceID {
				continue
			}
			GlobalChatHub.BroadcastLocal(chatPayload.RoomID, chatPayload.Event, chatPayload.Data, chatPayload.ExceptClientID)
		}
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
