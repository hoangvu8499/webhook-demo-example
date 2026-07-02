package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Kafka
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	// PostgreSQL
	DatabaseURL string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Concurrency
	WorkerCount int

	// HTTP delivery
	HTTPTimeout time.Duration

	// Retry scheduler
	RetryBaseDelay     time.Duration
	RetryCheckInterval time.Duration

	// Redis cache TTL
	CacheTTL time.Duration
}

func Load() *Config {
	return &Config{
		KafkaBrokers:       splitComma(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:         getEnv("KAFKA_TOPIC", "webhook-events"),
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "webhook-listener-group"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/webhook_db?pool_max_conns=30&pool_min_conns=5"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            envInt("REDIS_DB", 0),
		WorkerCount:        envInt("WORKER_COUNT", 200),
		HTTPTimeout:        envDuration("HTTP_TIMEOUT", 10*time.Second),
		RetryBaseDelay:     envDuration("RETRY_BASE_DELAY", 60*time.Second),
		RetryCheckInterval: envDuration("RETRY_CHECK_INTERVAL", 30*time.Second),
		CacheTTL:           envDuration("CACHE_TTL", 10*time.Minute),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
