package com.demo.webhook.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

/**
 * Generic webhook event payload matching Flodesk webhook event schema.
 * Covers: subscriber.created, subscriber.unsubscribed, subscriber.added_to_segment
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public class WebhookEvent {

    @JsonProperty("event_name")
    private String eventName;

    @JsonProperty("event_time")
    private String eventTime;

    private SubscriberInfo subscriber;

    @JsonProperty("webhook_id")
    private String webhookId;

    // Only present for subscriber.added_to_segment
    private Map<String, Object> segment;

    public String getEventName() { return eventName; }
    public void setEventName(String eventName) { this.eventName = eventName; }

    public String getEventTime() { return eventTime; }
    public void setEventTime(String eventTime) { this.eventTime = eventTime; }

    public SubscriberInfo getSubscriber() { return subscriber; }
    public void setSubscriber(SubscriberInfo subscriber) { this.subscriber = subscriber; }

    public String getWebhookId() { return webhookId; }
    public void setWebhookId(String webhookId) { this.webhookId = webhookId; }

    public Map<String, Object> getSegment() { return segment; }
    public void setSegment(Map<String, Object> segment) { this.segment = segment; }
}
