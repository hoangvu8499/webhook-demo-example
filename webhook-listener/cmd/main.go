package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"webhook-listener/internal/cache"
	"webhook-listener/internal/config"
	"webhook-listener/internal/consumer"
	"webhook-listener/internal/httpclient"
	"webhook-listener/internal/repository"
	"webhook-listener/internal/service"
)

func main() {
	// Load .env if present; ignore error when file doesn't exist (env vars pre-set by shell).
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("[WARN] .env load: %v", err)
	}

	cfg := config.Load()

	log.Printf("=== webhook-listener starting ===")
	log.Printf("Kafka brokers : %v", cfg.KafkaBrokers)
	log.Printf("Topic         : %s", cfg.KafkaTopic)
	log.Printf("Consumer group: %s", cfg.KafkaGroupID)
	log.Printf("Worker count  : %d", cfg.WorkerCount)
	log.Printf("Retry interval: %s", cfg.RetryCheckInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("PostgreSQL connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("PostgreSQL ping: %v", err)
	}
	log.Println("PostgreSQL connected")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisCache := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.CacheTTL)
	defer redisCache.Close()

	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("[WARN] Redis unavailable (%v) — cache disabled, all lookups hit DB", err)
	} else {
		log.Println("Redis connected")
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	webhookRepo := repository.NewWebhookRepo(pool)
	eventRepo := repository.NewEventRepo(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	httpClient := httpclient.New(cfg.HTTPTimeout)
	webhookSvc := service.NewWebhookService(redisCache, webhookRepo)
	dispatcher := service.NewDispatcher(webhookSvc, httpClient, eventRepo, cfg)

	// ── Kafka producer (retry scheduler re-queues events via this) ───────────
	kafkaProducer, err := consumer.NewKafkaProducer(cfg)
	if err != nil {
		log.Fatalf("Kafka producer init: %v", err)
	}
	defer kafkaProducer.Close()

	// ── Retry scheduler ───────────────────────────────────────────────────────
	retrySched := consumer.NewRetryScheduler(eventRepo, kafkaProducer, webhookSvc, cfg)
	go retrySched.Run(ctx)

	// ── Kafka consumer group ──────────────────────────────────────────────────
	kafkaConsumer, err := consumer.NewKafkaConsumer(cfg, dispatcher)
	if err != nil {
		log.Fatalf("Kafka consumer init: %v", err)
	}
	defer kafkaConsumer.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafkaConsumer.Run(ctx)
	}()

	log.Printf("=== webhook-listener ready — listening on topic=%s ===", cfg.KafkaTopic)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutdown signal received — draining workers...")
	cancel()
	wg.Wait() // block until Run() returns (which calls handler.drain())
	log.Println("Done.")
}
