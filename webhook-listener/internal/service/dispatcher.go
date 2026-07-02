package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"webhook-listener/internal/config"
	"webhook-listener/internal/httpclient"
	"webhook-listener/internal/model"
	"webhook-listener/internal/repository"
)

type Dispatcher struct {
	webhookSvc *WebhookService
	http       *httpclient.Client
	eventRepo  *repository.EventRepo
	cfg        *config.Config
}

func NewDispatcher(
	ws *WebhookService,
	h *httpclient.Client,
	er *repository.EventRepo,
	cfg *config.Config,
) *Dispatcher {
	return &Dispatcher{webhookSvc: ws, http: h, eventRepo: er, cfg: cfg}
}

// Process decodes a Kafka message, resolves the target URL, then delivers.
func (d *Dispatcher) Process(ctx context.Context, rawMsg []byte) {
	var event model.KafkaEvent
	if err := json.Unmarshal(rawMsg, &event); err != nil {
		log.Printf("[ERROR] parse kafka message: %v | raw=%s", err, truncate(string(rawMsg), 200))
		return
	}
	if event.WebhookID == "" || event.Subscriber == nil || event.Subscriber.ID == "" {
		log.Printf("[WARN] incomplete event: webhook_id=%s", event.WebhookID)
		return
	}

	cfg, err := d.webhookSvc.GetConfig(ctx, event.WebhookID)
	if err != nil {
		log.Printf("[ERROR] lookup webhook %s: %v", event.WebhookID, err)
		return
	}

	d.deliver(ctx, cfg, string(rawMsg), event.EventName, event.Subscriber.ID)
}

// deliver handles event delivery:
//  1. HTTP POST to provider
//  2. UPDATE webhook_events status (RETURNING id)
//  3. INSERT into webhook_event_logs
func (d *Dispatcher) deliver(
	ctx context.Context,
	cfg *model.WebhookConfig,
	payload, eventName, subscriberID string,
) {
	start := time.Now()
	result, httpErr := d.http.Post(ctx, cfg.PostURL, payload)
	durationMs := int(time.Since(start).Milliseconds())

	success := httpErr == nil && result.HTTPStatus >= 200 && result.HTTPStatus < 300

	var eventID string
	if success {
		id, err := d.eventRepo.MarkSuccess(ctx, cfg.ID, eventName, subscriberID)
		if err != nil {
			log.Printf("[WARN] mark_success webhook=%s sub=%s: %v", cfg.ID, subscriberID, err)
		}
		eventID = id
		log.Printf("[DELIVERED] webhook=%s event=%s http=%d ms=%d", cfg.ID, eventName, result.HTTPStatus, durationMs)
	} else {
		id, err := d.eventRepo.MarkRetry(ctx, cfg.ID, eventName, subscriberID, errStr(httpErr))
		if err != nil {
			log.Printf("[WARN] mark_retry webhook=%s sub=%s: %v", cfg.ID, subscriberID, err)
		}
		eventID = id
		log.Printf("[FAILED] webhook=%s event=%s http=%d err=%v ms=%d", cfg.ID, eventName, result.HTTPStatus, httpErr, durationMs)
	}

	if eventID != "" {
		d.writeLog(ctx, repository.EventLogEntry{
			EventID:        eventID,
			AttemptNumber:  1,
			HTTPStatusCode: result.HTTPStatus,
			ResponseBody:   result.Body,
			DurationMs:     durationMs,
			ErrorMessage:   errStr(httpErr),
		})
	}
}

func (d *Dispatcher) writeLog(ctx context.Context, e repository.EventLogEntry) {
	if err := d.eventRepo.InsertLog(ctx, e); err != nil {
		log.Printf("[ERROR] insert_log event=%s: %v", e.EventID, err)
	}
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
