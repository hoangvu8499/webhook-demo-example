package com.example.webhookgenerator.model;

import java.time.OffsetDateTime;
import java.util.UUID;

public class WebhookEventRecord {

    private final UUID id;
    private final UUID accountId;
    private final UUID webhookId;
    private final String postUrl;
    private final String eventName;
    private final OffsetDateTime eventTime;
    private final String payload;   // JSON string
    private final String dedupKey;

    public WebhookEventRecord(UUID id, UUID accountId, UUID webhookId, String postUrl,
                              String eventName, OffsetDateTime eventTime,
                              String payload, String dedupKey) {
        this.id = id;
        this.accountId = accountId;
        this.webhookId = webhookId;
        this.postUrl = postUrl;
        this.eventName = eventName;
        this.eventTime = eventTime;
        this.payload = payload;
        this.dedupKey = dedupKey;
    }

    public UUID getId() { return id; }
    public UUID getAccountId() { return accountId; }
    public UUID getWebhookId() { return webhookId; }
    public String getPostUrl() { return postUrl; }
    public String getEventName() { return eventName; }
    public OffsetDateTime getEventTime() { return eventTime; }
    public String getPayload() { return payload; }
    public String getDedupKey() { return dedupKey; }
}
