package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"webhook-listener/internal/model"
)

type WebhookRepo struct{ db *pgxpool.Pool }

func NewWebhookRepo(db *pgxpool.Pool) *WebhookRepo { return &WebhookRepo{db: db} }

// MarkInactive sets webhooks.status = 2 (inactive) for a given webhook.
func (r *WebhookRepo) MarkInactive(ctx context.Context, webhookID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE webhooks SET status = 2, updated_at = NOW() WHERE id = $1`,
		webhookID,
	)
	return err
}

// GetByID fetches an active webhook config from the webhooks table.
func (r *WebhookRepo) GetByID(ctx context.Context, webhookID string) (*model.WebhookConfig, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, account_id, post_url, events, status
		 FROM webhooks
		 WHERE id = $1 AND status = 1
		 LIMIT 1`,
		webhookID,
	)
	var cfg model.WebhookConfig
	if err := row.Scan(&cfg.ID, &cfg.AccountID, &cfg.PostURL, &cfg.Events, &cfg.Status); err != nil {
		return nil, fmt.Errorf("webhook not found id=%s: %w", webhookID, err)
	}
	return &cfg, nil
}
