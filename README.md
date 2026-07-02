# webhook-demo-example


webhook-event-generator: Simulates an event generator (instead of a real business logic system)

webhook-listener: Core component: consumes events, sends webhooks, retry, manages state

webhook-provider-demo: Simulates a client-side server (where the webhook is received), with different stability levels for retry testing

Technology used:
Java 8, Spring Boot framework → for module mock data and api

Go 1.22 → for Webhook-listener

Kafka, Redis, Postgres DB

