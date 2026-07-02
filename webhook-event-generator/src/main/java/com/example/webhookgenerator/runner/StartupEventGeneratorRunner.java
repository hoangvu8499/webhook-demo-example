package com.example.webhookgenerator.runner;

import com.example.webhookgenerator.service.EventGeneratorService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

@Component
public class StartupEventGeneratorRunner implements ApplicationRunner {

    private static final Logger log = LoggerFactory.getLogger(StartupEventGeneratorRunner.class);

    private final EventGeneratorService eventGeneratorService;

    public StartupEventGeneratorRunner(EventGeneratorService eventGeneratorService) {
        this.eventGeneratorService = eventGeneratorService;
    }

    @Override
    public void run(ApplicationArguments args) throws Exception {
        log.info("========================================");
        log.info("  Webhook Event Generator - STARTUP RUN ");
        log.info("========================================");
        log.info("Generating initial event batch for all accounts...");

        long wallStart = System.currentTimeMillis();
        eventGeneratorService.generateAll();
        long wallElapsed = System.currentTimeMillis() - wallStart;

        log.info("Startup event generation completed in {}ms", wallElapsed);
        log.info("Application will now exit. Re-run to generate another batch.");
    }
}
