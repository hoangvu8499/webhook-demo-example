package com.example.webhookgenerator.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@ConfigurationProperties(prefix = "generator")
public class GeneratorProperties {

    private int batchSize = 500;
    private List<AccountConfig> accounts = new ArrayList<>();

    public int getBatchSize() { return batchSize; }
    public void setBatchSize(int batchSize) { this.batchSize = batchSize; }

    public List<AccountConfig> getAccounts() { return accounts; }
    public void setAccounts(List<AccountConfig> accounts) { this.accounts = accounts; }

    public static class AccountConfig {
        private String id;
        private int eventCount;

        public String getId() { return id; }
        public void setId(String id) { this.id = id; }

        public int getEventCount() { return eventCount; }
        public void setEventCount(int eventCount) { this.eventCount = eventCount; }
    }
}
