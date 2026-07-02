package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"webhook-listener/internal/model"
)

// EventLogEntry maps to the actual webhook_event_logs schema.
type EventLogEntry struct {
	EventID        string // FK → webhook_events.id
	AttemptNumber  int
	HTTPStatusCode int
	ResponseBody   string
	ErrorMessage   string
	DurationMs     int
}

type EventRepo struct{ db *pgxpool.Pool }

func NewEventRepo(db *pgxpool.Pool) *EventRepo { return &EventRepo{db: db} }

// InsertLog records one delivery attempt in webhook_event_logs.
func (r *EventRepo) InsertLog(ctx context.Context, e EventLogEntry) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO webhook_event_logs
		 (event_id, attempt_number, http_status_code, response_body, error_message, duration_ms)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		e.EventID, e.AttemptNumber, e.HTTPStatusCode, e.ResponseBody, e.ErrorMessage, e.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	return nil
}

// MarkSuccess updates webhook_events to status=SUCCESS and returns the row's id for logging.
func (r *EventRepo) MarkSuccess(ctx context.Context, webhookID, eventName, subscriberID string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`UPDATE webhook_events
		 SET status            = $1,
		     attempt_count     = attempt_count + 1,
		     last_attempted_at = NOW(),
		     updated_at        = NOW()
		 WHERE webhook_id = $2
		   AND event_name = $3
		   AND payload->'subscriber'->>'id' = $4
		   AND status NOT IN ($1, $5)
		 RETURNING id`,
		model.StatusSuccess, webhookID, eventName, subscriberID, model.StatusFailedPermanently,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("mark success: %w", err)
	}
	return id, nil
}

// MarkRetry increments attempt_count, schedules next_retry_at (5s→10s→30s based on DB attempt_count),
// records last_error, and returns the row's id for logging.
func (r *EventRepo) MarkRetry(ctx context.Context, webhookID, eventName, subscriberID, lastError string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`UPDATE webhook_events
		 SET attempt_count     = attempt_count + 1,
		     next_retry_at     = NOW() + CASE
		                           WHEN attempt_count = 0 THEN interval '5 seconds'
		                           WHEN attempt_count = 1 THEN interval '10 seconds'
		                           ELSE interval '30 seconds'
		                         END,
		     last_error        = $1,
		     last_attempted_at = NOW(),
		     status            = $2,
		     updated_at        = NOW()
		 WHERE webhook_id = $3
		   AND event_name = $4
		   AND payload->'subscriber'->>'id' = $5
		   AND status NOT IN ($6, $7)
		 RETURNING id`,
		lastError, model.StatusPendingRetry,
		webhookID, eventName, subscriberID,
		model.StatusSuccess, model.StatusFailedPermanently,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("mark retry: %w", err)
	}
	return id, nil
}

// MarkDeadByID sets a specific webhook_events row to FAILED_PERMANENTLY.
// Called by RetryScheduler when attempt_count >= max_attempts.
func (r *EventRepo) MarkDeadByID(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE webhook_events
		 SET status            = $1,
		     last_attempted_at = NOW(),
		     updated_at        = NOW()
		 WHERE id = $2`,
		model.StatusFailedPermanently, id,
	)
	return err
}

// FetchPendingRetries atomically claims up to limit retry-due records.
// Sets next_retry_at = NOW() + 60s as a lease so the same record isn't
// claimed again if the worker crashes before updating it.
func (r *EventRepo) FetchPendingRetries(ctx context.Context, limit int) ([]model.RetryRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx,
		`SELECT id, webhook_id, account_id, post_url, event_name,
		        payload::text, attempt_count, max_attempts
		 FROM webhook_events
		 WHERE status = $1
		   AND next_retry_at <= NOW()
		 ORDER BY next_retry_at ASC
		 LIMIT $2
		 FOR UPDATE SKIP LOCKED`,
		model.StatusPendingRetry, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch retries: %w", err)
	}

	var recs []model.RetryRecord
	var ids []string
	for rows.Next() {
		var rec model.RetryRecord
		if err := rows.Scan(
			&rec.ID, &rec.WebhookID, &rec.AccountID, &rec.PostURL,
			&rec.EventName, &rec.Payload, &rec.AttemptCount, &rec.MaxAttempts,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan retry: %w", err)
		}
		recs = append(recs, rec)
		ids = append(ids, rec.ID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Lease: push next_retry_at forward to prevent double-claim if this worker crashes.
	for _, id := range ids {
		_, _ = tx.Exec(ctx,
			`UPDATE webhook_events SET next_retry_at = NOW() + interval '60 seconds' WHERE id = $1`,
			id,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit claim tx: %w", err)
	}
	return recs, nil
}
