package com.example.webhookgenerator;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class WebhookGeneratorApplication {

    public static void main(String[] args) {
        SpringApplication.run(WebhookGeneratorApplication.class, args);
    }
}
