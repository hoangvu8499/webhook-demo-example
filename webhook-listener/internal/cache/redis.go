package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"webhook-listener/internal/model"
)

const keyPrefix = "webhook:config:"

// RedisCache caches WebhookConfig by webhook_id with a configurable TTL.
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, db int, ttl time.Duration) *RedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     30,
		MinIdleConns: 5,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	return &RedisCache{client: c, ttl: ttl}
}

// Get returns cached WebhookConfig for webhookID; returns nil on cache miss.
func (r *RedisCache) Get(ctx context.Context, webhookID string) (*model.WebhookConfig, error) {
	data, err := r.client.Get(ctx, keyPrefix+webhookID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get %s: %w", webhookID, err)
	}
	var cfg model.WebhookConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("redis unmarshal: %w", err)
	}
	return &cfg, nil
}

// Set writes WebhookConfig into Redis with TTL. Errors are silently dropped
// so a Redis outage never blocks the critical delivery path.
func (r *RedisCache) Set(ctx context.Context, webhookID string, cfg *model.WebhookConfig) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	_ = r.client.Set(ctx, keyPrefix+webhookID, data, r.ttl).Err()
}

// Delete removes a cached webhook config so the next lookup hits the DB.
func (r *RedisCache) Delete(ctx context.Context, webhookID string) {
	_ = r.client.Del(ctx, keyPrefix+webhookID).Err()
}

func (r *RedisCache) Ping(ctx context.Context) error { return r.client.Ping(ctx).Err() }
func (r *RedisCache) Close() error                   { return r.client.Close() }
