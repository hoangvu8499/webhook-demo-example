package com.example.webhookgenerator.service;

import com.example.webhookgenerator.config.GeneratorProperties;
import com.example.webhookgenerator.model.WebhookEventRecord;
import com.example.webhookgenerator.model.WebhookInfo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.*;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class EventGeneratorService {

    private static final Logger log = LoggerFactory.getLogger(EventGeneratorService.class);

    private final GeneratorProperties properties;
    private final WebhookLookupService webhookLookupService;
    private final PayloadGeneratorService payloadGeneratorService;
    private final WebhookEventPersistenceService persistenceService;
    private final KafkaPublisherService kafkaPublisherService;

    public EventGeneratorService(GeneratorProperties properties,
                                 WebhookLookupService webhookLookupService,
                                 PayloadGeneratorService payloadGeneratorService,
                                 WebhookEventPersistenceService persistenceService,
                                 KafkaPublisherService kafkaPublisherService) {
        this.properties = properties;
        this.webhookLookupService = webhookLookupService;
        this.payloadGeneratorService = payloadGeneratorService;
        this.persistenceService = persistenceService;
        this.kafkaPublisherService = kafkaPublisherService;
    }

    // Processes accounts sequentially in config order: ALPHA → BETA → GAMMA
    public void generateAll() {
        List<GeneratorProperties.AccountConfig> accountConfigs = properties.getAccounts();
        if (accountConfigs.isEmpty()) {
            log.warn("No accounts configured for event generation.");
            return;
        }

        int totalInserted = 0;
        int totalPublished = 0;

        for (GeneratorProperties.AccountConfig cfg : accountConfigs) {
            UUID accountId = UUID.fromString(cfg.getId());
            WebhookInfo webhook = webhookLookupService.findByAccountId(accountId);
            log.info("Loaded webhook for account [{}]: webhook_id={}", cfg.getId(), webhook.getId());

            int count = cfg.getEventCount();
            long start = System.currentTimeMillis();
            log.info("Starting generation of {} events for account [{}]", count, accountId);

            int inserted = generateForAccount(webhook, count);
            totalInserted += inserted;
            totalPublished += count;

            long elapsed = System.currentTimeMillis() - start;
            log.info("Finished account [{}]: {} events inserted, {}ms", accountId, inserted, elapsed);
        }

        log.info("=== Generation complete: {} total events inserted to DB, {} queued to Kafka ===",
                totalInserted, totalPublished);
    }

    private int generateForAccount(WebhookInfo webhook, int totalCount) {
        int batchSize = properties.getBatchSize();
        int inserted = 0;
        int remaining = totalCount;

        while (remaining > 0) {
            int currentBatch = Math.min(remaining, batchSize);
            List<WebhookEventRecord> batch = buildBatch(webhook, currentBatch);

            int savedCount = persistenceService.batchInsert(batch);
            inserted += savedCount;

            // Publish saved events to Kafka (fire-and-forget, async callbacks)
            kafkaPublisherService.publishAll(batch);

            remaining -= currentBatch;
            log.debug("Account [{}]: saved {}/{}, remaining {}",
                    webhook.getAccountId(), inserted, totalCount, remaining);
        }

        return inserted;
    }

    private List<WebhookEventRecord> buildBatch(WebhookInfo webhook, int count) {
        List<String> supportedEvents = webhook.getEvents();
        List<WebhookEventRecord> batch = new ArrayList<>(count);

        for (int i = 0; i < count; i++) {
            UUID eventId = UUID.randomUUID();
            String eventName = supportedEvents.get(
                    ThreadLocalRandom.current().nextInt(supportedEvents.size())
            );
            OffsetDateTime eventTime = OffsetDateTime.now();
            String payload = payloadGeneratorService.generatePayload(eventName, webhook.getId().toString());
            // dedup_key is unique per generated event; using eventId ensures no collisions
            String dedupKey = eventId.toString();

            batch.add(new WebhookEventRecord(
                    eventId,
                    webhook.getAccountId(),
                    webhook.getId(),
                    webhook.getPostUrl(),
                    eventName,
                    eventTime,
                    payload,
                    dedupKey
            ));
        }

        return batch;
    }
}
