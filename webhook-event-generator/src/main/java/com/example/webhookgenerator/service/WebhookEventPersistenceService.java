package com.example.webhookgenerator.service;

import com.example.webhookgenerator.model.WebhookEventRecord;
import org.postgresql.util.PGobject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.sql.PreparedStatement;
import java.sql.SQLException;
import java.sql.Timestamp;
import java.util.List;
import java.util.UUID;

@Service
public class WebhookEventPersistenceService {

    private static final Logger log = LoggerFactory.getLogger(WebhookEventPersistenceService.class);

    private static final String INSERT_SQL =
            "INSERT INTO webhook_events " +
            "(id, account_id, webhook_id, post_url, event_name, event_time, payload, dedup_key, " +
            " status, attempt_count, max_attempts, created_at, updated_at) " +
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 3, NOW(), NOW()) " +
            "ON CONFLICT (dedup_key) DO NOTHING";

    private final JdbcTemplate jdbcTemplate;

    public WebhookEventPersistenceService(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public int batchInsert(List<WebhookEventRecord> events) {
        int[][] result = jdbcTemplate.batchUpdate(INSERT_SQL, events, events.size(),
                (PreparedStatement ps, WebhookEventRecord event) -> {
                    ps.setObject(1, event.getId());
                    ps.setObject(2, event.getAccountId());
                    ps.setObject(3, event.getWebhookId());
                    ps.setString(4, event.getPostUrl());
                    ps.setString(5, event.getEventName());
                    ps.setTimestamp(6, Timestamp.from(event.getEventTime().toInstant()));
                    ps.setObject(7, buildJsonbObject(event.getPayload()));
                    ps.setString(8, event.getDedupKey());
                }
        );

        int total = 0;
        for (int[] batch : result) {
            for (int count : batch) {
                total += Math.max(count, 0);
            }
        }
        return total;
    }

    private PGobject buildJsonbObject(String json) {
        try {
            PGobject pgObject = new PGobject();
            pgObject.setType("jsonb");
            pgObject.setValue(json);
            return pgObject;
        } catch (SQLException e) {
            throw new RuntimeException("Failed to create JSONB object", e);
        }
    }
}
