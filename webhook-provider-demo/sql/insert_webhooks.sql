-- ============================================================
-- INSERT script: webhooks table
-- Project: webhook-provider-demo
-- Environment: localhost (port 8081)
-- Created: 2026-06-29
--
-- 3 webhooks tương ứng với 3 provider endpoints:
--   - provider-alpha : luôn SUCCESS (HTTP 200)
--   - provider-beta  : luôn FAIL    (HTTP 500)
--   - provider-gamma : ngẫu nhiên  (HTTP 200 | 500)
--
-- Mỗi webhook thuộc một account khác nhau để giả lập
-- 3 khách hàng độc lập đang dùng hệ thống.
-- ============================================================


-- ----------------------------------------------------------------
-- 1. Provider Alpha — reliable endpoint, always SUCCESS
--    Account: acc-alpha  |  Status: ACTIVE (1)
--    Đăng ký nhận toàn bộ 3 loại event
-- ----------------------------------------------------------------
INSERT INTO webhooks (
    id,
    account_id,
    post_url,
    events,
    status
) VALUES (
    'aaaaaaaa-0001-0001-0001-000000000001',
    'cccccccc-0001-0001-0001-000000000001',
    'http://localhost:8081/webhook/provider-alpha',
    ARRAY[
        'subscriber.created',
        'subscriber.unsubscribed',
        'subscriber.added_to_segment'
    ],
    1  -- ACTIVE
);


-- ----------------------------------------------------------------
-- 2. Provider Beta — broken endpoint, always FAIL
--    Account: acc-beta   |  Status: ACTIVE (1)
--    Đăng ký nhận toàn bộ 3 loại event
--    → Dùng để test retry / failed_permanently
-- ----------------------------------------------------------------
INSERT INTO webhooks (
    id,
    account_id,
    post_url,
    events,
    status
) VALUES (
    'bbbbbbbb-0002-0002-0002-000000000002',
    'cccccccc-0002-0002-0002-000000000002',
    'http://localhost:8081/webhook/provider-beta',
    ARRAY[
        'subscriber.created',
        'subscriber.unsubscribed',
        'subscriber.added_to_segment'
    ],
    1  -- ACTIVE
);


-- ----------------------------------------------------------------
-- 3. Provider Gamma — flaky endpoint, random SUCCESS/FAIL
--    Account: acc-gamma  |  Status: ACTIVE (1)
--    Đăng ký nhận toàn bộ 3 loại event
--    → Dùng để test retry + exponential backoff
-- ----------------------------------------------------------------
INSERT INTO webhooks (
    id,
    account_id,
    post_url,
    events,
    status
) VALUES (
    'cccccccc-0003-0003-0003-000000000003',
    'cccccccc-0003-0003-0003-000000000003',
    'http://localhost:8081/webhook/provider-gamma',
    ARRAY[
        'subscriber.created',
        'subscriber.unsubscribed',
        'subscriber.added_to_segment'
    ],
    1  -- ACTIVE
);


-- ----------------------------------------------------------------
-- Verify
-- ----------------------------------------------------------------
SELECT
    id,
    account_id,
    post_url,
    events,
    status,
    created_at
FROM webhooks
ORDER BY created_at;
