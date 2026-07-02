package com.example.webhookgenerator.config;

import com.example.webhookgenerator.service.PartitionConfigService;
import org.apache.kafka.clients.admin.NewTopic;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.config.TopicBuilder;

@Configuration
public class KafkaTopicConfig {

    private static final Logger log = LoggerFactory.getLogger(KafkaTopicConfig.class);

    @Value("${kafka.topic.name}")
    private String topicName;

    @Value("${kafka.topic.replication-factor}")
    private short replicationFactor;

    private final PartitionConfigService partitionConfigService;

    public KafkaTopicConfig(PartitionConfigService partitionConfigService) {
        this.partitionConfigService = partitionConfigService;
    }

    @Bean
    public NewTopic webhookEventsTopic() {
        int partitions = partitionConfigService.getPartitionCount();
        log.info("Creating Kafka topic '{}' with {} partitions", topicName, partitions);
        return TopicBuilder.name(topicName)
                .partitions(partitions)
                .replicas(replicationFactor)
                .build();
    }
}
