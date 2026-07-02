package com.demo.webhook.controller;

import com.demo.webhook.model.WebhookEvent;
import com.demo.webhook.model.WebhookResponse;
import com.demo.webhook.model.WebhookResponse.DeliveryStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.Instant;
import java.util.Random;

/**
 * Simulates 3 third-party webhook receiver endpoints with different reliability behaviors.
 *
 * Provider A — always succeeds (HTTP 200)
 * Provider B — always fails (HTTP 500)
 * Provider C — randomly succeeds or fails (50/50)
 */
@RestController
@RequestMapping("/webhook")
public class WebhookReceiverController {

    private static final Logger log = LoggerFactory.getLogger(WebhookReceiverController.class);

    private static final String PROVIDER_A = "provider-alpha";
    private static final String PROVIDER_B = "provider-beta";
    private static final String PROVIDER_C = "provider-gamma";

    private final Random random = new Random();

    /**
     * Provider Alpha — always returns SUCCESS (HTTP 200).
     * Simulates a reliable, well-behaved third-party endpoint.
     */
    @PostMapping("/provider-alpha")
    public ResponseEntity<WebhookResponse> providerAlpha(@RequestBody WebhookEvent event) {
        log.info("[{}] Received event: {} | webhook_id: {}", PROVIDER_A, event.getEventName(), event.getWebhookId());

        WebhookResponse response = new WebhookResponse(
                DeliveryStatus.SUCCESS,
                "Event received and processed successfully.",
                PROVIDER_A,
                event.getEventName(),
                Instant.now().toString()
        );

        return ResponseEntity.ok(response);
    }

    /**
     * Provider Beta — always returns FAIL (HTTP 500).
     * Simulates an unreliable or broken third-party endpoint.
     */
    @PostMapping("/provider-beta")
    public ResponseEntity<WebhookResponse> providerBeta(@RequestBody WebhookEvent event) {
        log.warn("[{}] Received event: {} | webhook_id: {} — responding with FAIL",
                PROVIDER_B, event.getEventName(), event.getWebhookId());

        WebhookResponse response = new WebhookResponse(
                DeliveryStatus.FAIL,
                "Internal server error: provider endpoint is unavailable.",
                PROVIDER_B,
                event.getEventName(),
                Instant.now().toString()
        );

        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
    }

    /**
     * Provider Gamma — randomly returns SUCCESS (HTTP 200) or FAIL (HTTP 500).
     * Simulates a flaky third-party endpoint with ~50% failure rate.
     * Useful for testing retry and backoff logic.
     */
    @PostMapping("/provider-gamma")
    public ResponseEntity<WebhookResponse> providerGamma(@RequestBody WebhookEvent event) {
        boolean success = random.nextBoolean();

        log.info("[{}] Received event: {} | webhook_id: {} — outcome: {}",
                PROVIDER_C, event.getEventName(), event.getWebhookId(), success ? "SUCCESS" : "FAIL");

        if (success) {
            WebhookResponse response = new WebhookResponse(
                    DeliveryStatus.SUCCESS,
                    "Event received and processed successfully.",
                    PROVIDER_C,
                    event.getEventName(),
                    Instant.now().toString()
            );
            return ResponseEntity.ok(response);
        } else {
            WebhookResponse response = new WebhookResponse(
                    DeliveryStatus.FAIL,
                    "Transient error: provider failed to process the event. Please retry.",
                    PROVIDER_C,
                    event.getEventName(),
                    Instant.now().toString()
            );
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
        }
    }
}
