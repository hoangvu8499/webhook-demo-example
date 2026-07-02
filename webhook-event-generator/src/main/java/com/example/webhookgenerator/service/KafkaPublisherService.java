package com.example.webhookgenerator.service;

import com.example.webhookgenerator.model.WebhookEventRecord;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.SendResult;
import org.springframework.stereotype.Service;
import org.springframework.util.concurrent.ListenableFuture;
import org.springframework.util.concurrent.ListenableFutureCallback;

import java.util.List;

@Service
public class KafkaPublisherService {

    private static final Logger log = LoggerFactory.getLogger(KafkaPublisherService.class);

    @Value("${kafka.topic.name}")
    private String topicName;

    private final KafkaTemplate<String, String> kafkaTemplate;

    public KafkaPublisherService(KafkaTemplate<String, String> kafkaTemplate) {
        this.kafkaTemplate = kafkaTemplate;
    }

    public void publishAll(List<WebhookEventRecord> events) {
        for (WebhookEventRecord event : events) {
            publish(event);
        }
    }

    public void publish(WebhookEventRecord event) {
        String value = event.getPayload();

        ListenableFuture<SendResult<String, String>> future =
                kafkaTemplate.send(topicName, value);

        future.addCallback(new ListenableFutureCallback<SendResult<String, String>>() {
            @Override
            public void onSuccess(SendResult<String, String> result) {
                log.debug("Published event [{}] to partition {} offset {}",
                        event.getId(),
                        result.getRecordMetadata().partition(),
                        result.getRecordMetadata().offset());
            }

            @Override
            public void onFailure(Throwable ex) {
                log.error("Failed to publish event [{}] account [{}]: {}",
                        event.getId(), event.getAccountId(), ex.getMessage());
            }
        });
    }
}
