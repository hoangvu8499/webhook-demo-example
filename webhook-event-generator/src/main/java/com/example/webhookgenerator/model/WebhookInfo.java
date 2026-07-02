package com.example.webhookgenerator.model;

import java.util.List;
import java.util.UUID;

public class WebhookInfo {

    private final UUID id;
    private final UUID accountId;
    private final String postUrl;
    private final List<String> events;

    public WebhookInfo(UUID id, UUID accountId, String postUrl, List<String> events) {
        this.id = id;
        this.accountId = accountId;
        this.postUrl = postUrl;
        this.events = events;
    }

    public UUID getId() { return id; }
    public UUID getAccountId() { return accountId; }
    public String getPostUrl() { return postUrl; }
    public List<String> getEvents() { return events; }
}
