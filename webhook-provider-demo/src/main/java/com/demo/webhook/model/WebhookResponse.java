package com.demo.webhook.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public class WebhookResponse {

    public enum DeliveryStatus {
        SUCCESS, FAIL
    }

    private DeliveryStatus status;
    private String message;
    private String provider;

    @JsonProperty("received_event")
    private String receivedEvent;

    @JsonProperty("received_at")
    private String receivedAt;

    public WebhookResponse() {}

    public WebhookResponse(DeliveryStatus status, String message, String provider,
                           String receivedEvent, String receivedAt) {
        this.status = status;
        this.message = message;
        this.provider = provider;
        this.receivedEvent = receivedEvent;
        this.receivedAt = receivedAt;
    }

    public DeliveryStatus getStatus() { return status; }
    public void setStatus(DeliveryStatus status) { this.status = status; }

    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }

    public String getProvider() { return provider; }
    public void setProvider(String provider) { this.provider = provider; }

    public String getReceivedEvent() { return receivedEvent; }
    public void setReceivedEvent(String receivedEvent) { this.receivedEvent = receivedEvent; }

    public String getReceivedAt() { return receivedAt; }
    public void setReceivedAt(String receivedAt) { this.receivedAt = receivedAt; }
}
