package com.example.webhookgenerator.service;

import com.example.webhookgenerator.model.WebhookInfo;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.sql.Array;
import java.util.*;

@Service
public class WebhookLookupService {

    private static final List<String> DEFAULT_EVENTS = Arrays.asList(
            "subscriber.created", "subscriber.unsubscribed", "subscriber.added_to_segment"
    );

    private final JdbcTemplate jdbcTemplate;

    public WebhookLookupService(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public WebhookInfo findByAccountId(UUID accountId) {
        List<WebhookInfo> results = jdbcTemplate.query(
                "SELECT id, account_id, post_url, events FROM webhooks WHERE account_id = ?::uuid AND status = 1 LIMIT 1",
                (rs, rowNum) -> {
                    UUID id = UUID.fromString(rs.getString("id"));
                    UUID aid = UUID.fromString(rs.getString("account_id"));
                    String postUrl = rs.getString("post_url");

                    Array eventsArray = rs.getArray("events");
                    List<String> events;
                    if (eventsArray != null) {
                        String[] arr = (String[]) eventsArray.getArray();
                        events = Arrays.asList(arr);
                    } else {
                        events = DEFAULT_EVENTS;
                    }

                    return new WebhookInfo(id, aid, postUrl, events);
                },
                accountId.toString()
        );

        if (results.isEmpty()) {
            throw new IllegalStateException(
                    "No active webhook found for account_id=" + accountId
                            + ". Please ensure the webhooks table has a record for this account."
            );
        }

        return results.get(0);
    }

    public Map<UUID, WebhookInfo> findAll(List<UUID> accountIds) {
        Map<UUID, WebhookInfo> result = new LinkedHashMap<>();
        for (UUID accountId : accountIds) {
            result.put(accountId, findByAccountId(accountId));
        }
        return result;
    }
}
