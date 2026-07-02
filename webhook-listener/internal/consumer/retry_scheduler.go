package consumer

import (
	"context"
	"log"
	"time"

	"webhook-listener/internal/config"
	"webhook-listener/internal/repository"
	"webhook-listener/internal/service"
)

const retryBatchSize = 100

// RetryScheduler polls webhook_events for PENDING_RETRY records whose next_retry_at
// has passed, then either pushes them back to Kafka or marks them permanently failed.
type RetryScheduler struct {
	eventRepo  *repository.EventRepo
	publisher  *KafkaProducer
	webhookSvc *service.WebhookService
	interval   time.Duration
}

func NewRetryScheduler(
	er *repository.EventRepo,
	pub *KafkaProducer,
	ws *service.WebhookService,
	cfg *config.Config,
) *RetryScheduler {
	return &RetryScheduler{
		eventRepo:  er,
		publisher:  pub,
		webhookSvc: ws,
		interval:   cfg.RetryCheckInterval,
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (s *RetryScheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	log.Printf("[RETRY-SCHEDULER] started, interval=%s batch=%d", s.interval, retryBatchSize)

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			log.Println("[RETRY-SCHEDULER] stopped")
			return
		}
	}
}

func (s *RetryScheduler) tick(ctx context.Context) {
	records, err := s.eventRepo.FetchPendingRetries(ctx, retryBatchSize)
	if err != nil {
		log.Printf("[RETRY-SCHEDULER] fetch error: %v", err)
		return
	}
	if len(records) == 0 {
		return
	}

	log.Printf("[RETRY-SCHEDULER] processing %d records", len(records))

	for _, rec := range records {
		if rec.AttemptCount >= rec.MaxAttempts {
			// Đã retry đủ lần → đánh dấu dead, không push queue nữa
			if err := s.eventRepo.MarkDeadByID(ctx, rec.ID); err != nil {
				log.Printf("[RETRY-SCHEDULER] mark_dead id=%s: %v", rec.ID, err)
				continue
			}
			log.Printf("[RETRY-DEAD] id=%s webhook=%s attempts=%d/%d", rec.ID, rec.WebhookID, rec.AttemptCount, rec.MaxAttempts)
			s.webhookSvc.MarkInactive(ctx, rec.WebhookID)
			continue
		}

		// Kiểm tra webhook còn active không trước khi push queue
		if _, err := s.webhookSvc.GetConfig(ctx, rec.WebhookID); err != nil {
			if err2 := s.eventRepo.MarkDeadByID(ctx, rec.ID); err2 != nil {
				log.Printf("[RETRY-SCHEDULER] mark_dead id=%s: %v", rec.ID, err2)
				continue
			}
			log.Printf("[RETRY-DEAD] id=%s webhook=%s (webhook not found/inactive)", rec.ID, rec.WebhookID)
			continue
		}

		// Còn lượt retry → push lại vào Kafka queue
		if err := s.publisher.Publish(rec.Payload); err != nil {
			log.Printf("[RETRY-SCHEDULER] publish id=%s: %v", rec.ID, err)
			continue
		}
		log.Printf("[RETRY-QUEUED] id=%s webhook=%s event=%s attempt=%d/%d", rec.ID, rec.WebhookID, rec.EventName, rec.AttemptCount+1, rec.MaxAttempts)
	}
}
