package service

import (
	"context"
	"fmt"
	"log"

	"webhook-listener/internal/cache"
	"webhook-listener/internal/model"
	"webhook-listener/internal/repository"
)

// WebhookService resolves webhook configuration with a Redis-first, DB-fallback strategy.
type WebhookService struct {
	cache *cache.RedisCache
	repo  *repository.WebhookRepo
}

func NewWebhookService(c *cache.RedisCache, r *repository.WebhookRepo) *WebhookService {
	return &WebhookService{cache: c, repo: r}
}

// MarkInactive sets the webhook to inactive in DB and removes it from Redis cache.
func (s *WebhookService) MarkInactive(ctx context.Context, webhookID string) {
	if err := s.repo.MarkInactive(ctx, webhookID); err != nil {
		log.Printf("[ERROR] mark_webhook_inactive webhook=%s: %v", webhookID, err)
		return
	}
	s.cache.Delete(ctx, webhookID)
	log.Printf("[WEBHOOK-INACTIVE] webhook=%s set to inactive", webhookID)
}

// GetConfig looks up webhook config by webhookID:
//  1. Redis cache hit → return immediately.
//  2. Cache miss     → query webhooks table, then cache the result.
func (s *WebhookService) GetConfig(ctx context.Context, webhookID string) (*model.WebhookConfig, error) {
	// 1. Redis
	cfg, err := s.cache.Get(ctx, webhookID)
	if err != nil {
		log.Printf("[WARN] redis get webhook %s: %v", webhookID, err)
	}
	if cfg != nil {
		return cfg, nil
	}

	// 2. DB
	cfg, err = s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("webhook lookup %s: %w", webhookID, err)
	}

	// 3. Write-through cache
	s.cache.Set(ctx, webhookID, cfg)
	return cfg, nil
}
