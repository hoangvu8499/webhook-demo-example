-- ============================================================
-- webhook-listener — SQL schema
-- Compatible with PostgreSQL 13+
-- Run this ONCE before starting the listener service.
-- ============================================================

-- ── webhook_event_logs ────────────────────────────────────────────────────────
-- Stores every delivery attempt (SUCCESS or FAIL) for auditing and debugging.
-- The generator already creates webhook_events and webhooks; this table is new.

CREATE TABLE IF NOT EXISTS webhook_event_logs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id     UUID        NOT NULL,
    account_id     UUID,
    post_url       TEXT        NOT NULL,
    event_name     VARCHAR(100) NOT NULL,
    payload        JSONB,                          -- event payload sent to provider
    http_status    INTEGER,                        -- HTTP status returned by provider
    status         VARCHAR(20) NOT NULL,           -- 'SUCCESS' or 'FAIL'
    response_body  TEXT,                           -- raw response body (truncated)
    attempt_number INTEGER     NOT NULL DEFAULT 1, -- 1 = first attempt, 2 = first retry …
    error_message  TEXT,                           -- network/timeout error if any
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wel_webhook_id   ON webhook_event_logs (webhook_id);
CREATE INDEX IF NOT EXISTS idx_wel_created_at   ON webhook_event_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wel_status       ON webhook_event_logs (status);

-- ── webhook_events (generator already creates this; verify columns) ──────────
-- These columns MUST exist.  Run the ALTER statements only if they are missing.

ALTER TABLE webhook_events
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_we_retry
    ON webhook_events (status, next_retry_at)
    WHERE status = 2;           -- only index PENDING_RETRY rows

-- ── Verify ───────────────────────────────────────────────────────────────────
SELECT
    table_name,
    column_name,
    data_type
FROM information_schema.columns
WHERE table_name IN ('webhook_events', 'webhook_event_logs', 'webhooks')
ORDER BY table_name, ordinal_position;
