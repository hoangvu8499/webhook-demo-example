package com.example.webhookgenerator.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class PayloadGeneratorService {

    private static final String[] FIRST_NAMES = {
        "Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry",
        "Iris", "Jack", "Karen", "Leo", "Mia", "Noah", "Olivia", "Paul",
        "Quinn", "Rachel", "Sam", "Tina", "Uma", "Victor", "Wendy", "Xander",
        "Yara", "Zoe"
    };

    private static final String[] LAST_NAMES = {
        "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
        "Davis", "Wilson", "Anderson", "Taylor", "Thomas", "Jackson", "White",
        "Harris", "Martin", "Thompson", "Young", "Robinson", "Clark"
    };

    private static final String[] EMAIL_DOMAINS = {
        "gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "example.com",
        "mail.com", "protonmail.com", "icloud.com"
    };

    private static final String[] SOURCES = {
        "manual", "csv", "form_optin", "integration", "checkout"
    };

    private static final String[] SUBSCRIBER_STATUSES = {
        "active", "active", "active", "unsubscribed", "unconfirmed"
    };

    private static final String[] SEGMENT_NAMES = {
        "VIP Customers", "Newsletter Subscribers", "New Users", "Premium Members",
        "Beta Testers", "Active Buyers", "Trial Users", "Loyal Customers"
    };

    private final ObjectMapper objectMapper;

    public PayloadGeneratorService(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public String generatePayload(String eventName, String webhookId) {
        try {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("event_name", eventName);
            payload.put("event_time", OffsetDateTime.now().format(DateTimeFormatter.ISO_OFFSET_DATE_TIME));
            payload.put("webhook_id", webhookId);
            payload.put("subscriber", buildSubscriberData());

            if ("subscriber.added_to_segment".equals(eventName)) {
                payload.put("segment", buildSegmentData());
            }

            return objectMapper.writeValueAsString(payload);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Failed to serialize event payload", e);
        }
    }

    private Map<String, Object> buildSubscriberData() {
        ThreadLocalRandom rng = ThreadLocalRandom.current();
        String firstName = FIRST_NAMES[rng.nextInt(FIRST_NAMES.length)];
        String lastName = LAST_NAMES[rng.nextInt(LAST_NAMES.length)];
        String domain = EMAIL_DOMAINS[rng.nextInt(EMAIL_DOMAINS.length)];
        String email = firstName.toLowerCase() + "." + lastName.toLowerCase()
                + rng.nextInt(10000) + "@" + domain;

        Map<String, Object> subscriber = new LinkedHashMap<>();
        subscriber.put("id", UUID.randomUUID().toString());
        subscriber.put("status", SUBSCRIBER_STATUSES[rng.nextInt(SUBSCRIBER_STATUSES.length)]);
        subscriber.put("email", email);
        subscriber.put("source", SOURCES[rng.nextInt(SOURCES.length)]);
        subscriber.put("first_name", firstName);
        subscriber.put("last_name", lastName);
        subscriber.put("segments", Collections.emptyList());
        subscriber.put("custom_fields", Collections.emptyMap());
        subscriber.put("optin_ip", randomIp(rng));
        subscriber.put("optin_timestamp", OffsetDateTime.now().minusDays(rng.nextInt(365))
                .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME));
        subscriber.put("created_at", OffsetDateTime.now().minusDays(rng.nextInt(365))
                .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME));
        return subscriber;
    }

    private Map<String, Object> buildSegmentData() {
        ThreadLocalRandom rng = ThreadLocalRandom.current();
        Map<String, Object> segment = new LinkedHashMap<>();
        segment.put("id", UUID.randomUUID().toString());
        segment.put("name", SEGMENT_NAMES[rng.nextInt(SEGMENT_NAMES.length)]);
        return segment;
    }

    private String randomIp(ThreadLocalRandom rng) {
        return rng.nextInt(256) + "." + rng.nextInt(256) + "."
                + rng.nextInt(256) + "." + rng.nextInt(256);
    }
}
