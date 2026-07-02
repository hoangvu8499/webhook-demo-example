package com.example.webhookgenerator.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class PartitionConfigService {

    private static final Logger log = LoggerFactory.getLogger(PartitionConfigService.class);
    private static final String REDIS_KEY = "kafka:config:partition_number";
    private static final int DEFAULT_PARTITIONS = 50;

    private final StringRedisTemplate redisTemplate;
    private final JdbcTemplate jdbcTemplate;

    public PartitionConfigService(StringRedisTemplate redisTemplate, JdbcTemplate jdbcTemplate) {
        this.redisTemplate = redisTemplate;
        this.jdbcTemplate = jdbcTemplate;
    }

    public int getPartitionCount() {
        // 1. Try Redis
        try {
            String cached = redisTemplate.opsForValue().get(REDIS_KEY);
            if (cached != null) {
                log.info("Partition count loaded from Redis: {}", cached);
                return Integer.parseInt(cached);
            }
        } catch (Exception e) {
            log.warn("Redis unavailable, falling back to DB: {}", e.getMessage());
        }

        // 2. Try DB
        try {
            List<Integer> results = jdbcTemplate.query(
                    "SELECT param_value FROM app_config WHERE param_group='KAFKA' AND param_name='PARTITION_NUMBER' AND status=1",
                    (rs, rowNum) -> Integer.parseInt(rs.getString("param_value"))
            );
            if (!results.isEmpty()) {
                int count = results.get(0);
                log.info("Partition count loaded from DB: {}", count);
                cacheToRedis(count);
                return count;
            }
        } catch (Exception e) {
            log.warn("Failed to fetch partition count from DB: {}", e.getMessage());
        }

        // 3. Default
        log.info("Partition count using default: {}", DEFAULT_PARTITIONS);
        cacheToRedis(DEFAULT_PARTITIONS);
        return DEFAULT_PARTITIONS;
    }

    private void cacheToRedis(int count) {
        try {
            redisTemplate.opsForValue().set(REDIS_KEY, String.valueOf(count));
        } catch (Exception e) {
            log.warn("Failed to cache partition count to Redis: {}", e.getMessage());
        }
    }
}
