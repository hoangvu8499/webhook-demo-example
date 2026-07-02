# webhook-provider-demo

## Project Overview

Mock server that simulates **3 third-party webhook receiver endpoints** with different reliability behaviors. Built to support the Webhook Notifier system (BE Home Assignment), where the main notifier service sends Flodesk webhook events to external providers.

This is **Giai đoạn 2 (Mock server D)** from `webhook-notifier-plan.md`.

## Stack

- **Java 8**
- **Spring Boot 2.7.18** (Spring Framework 5.3.x)
- **Maven**
- Port: `8081` (intentionally different from the main notifier service)

## Endpoints

All endpoints accept `POST` with `Content-Type: application/json`.

| Provider | URL | Behavior |
|---|---|---|
| Provider Alpha | `POST /webhook/provider-alpha` | Always returns **HTTP 200** (SUCCESS) |
| Provider Beta | `POST /webhook/provider-beta` | Always returns **HTTP 500** (FAIL) |
| Provider Gamma | `POST /webhook/provider-gamma` | Random **HTTP 200 or 500** (~50/50) |

## Request Body

Follows the Flodesk webhook event schema from `openapi.json`. All 3 event types are accepted:

**subscriber.created / subscriber.unsubscribed:**
```json
{
  "event_name": "subscriber.created",
  "event_time": "2024-01-02T15:04:05.999Z",
  "webhook_id": "wh_abc123",
  "subscriber": {
    "id": "sub_001",
    "email": "user@example.com",
    "status": "active",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

**subscriber.added_to_segment** (includes `segment` field):
```json
{
  "event_name": "subscriber.added_to_segment",
  "event_time": "2024-01-02T15:04:05.999Z",
  "webhook_id": "wh_abc123",
  "subscriber": { "id": "sub_001", "email": "user@example.com" },
  "segment": { "id": "seg_001", "name": "VIP Customers" }
}
```

## Response Body

```json
{
  "status": "SUCCESS",
  "message": "Event received and processed successfully.",
  "provider": "provider-alpha",
  "received_event": "subscriber.created",
  "received_at": "2024-01-02T15:04:05.999Z"
}
```

## How to Run

```bash
# From project root
mvn spring-boot:run

# Or build and run JAR
mvn clean package
java -jar target/webhook-provider-demo-1.0.0.jar
```

## How to Test (curl examples)

```bash
# Provider Alpha — always SUCCESS
curl -s -X POST http://localhost:8081/webhook/provider-alpha \
  -H "Content-Type: application/json" \
  -d '{"event_name":"subscriber.created","event_time":"2024-01-01T00:00:00Z","webhook_id":"wh_001","subscriber":{"id":"sub_001","email":"test@example.com","status":"active"}}'

# Provider Beta — always FAIL
curl -s -X POST http://localhost:8081/webhook/provider-beta \
  -H "Content-Type: application/json" \
  -d '{"event_name":"subscriber.unsubscribed","event_time":"2024-01-01T00:00:00Z","webhook_id":"wh_001","subscriber":{"id":"sub_001","email":"test@example.com"}}'

# Provider Gamma — random outcome
curl -s -X POST http://localhost:8081/webhook/provider-gamma \
  -H "Content-Type: application/json" \
  -d '{"event_name":"subscriber.added_to_segment","event_time":"2024-01-01T00:00:00Z","webhook_id":"wh_001","subscriber":{"id":"sub_001","email":"test@example.com"},"segment":{"id":"seg_001","name":"VIP"}}'
```

## Project Structure

```
src/main/java/com/demo/webhook/
├── WebhookProviderDemoApplication.java   # Spring Boot entry point
├── controller/
│   └── WebhookReceiverController.java   # 3 POST endpoints
└── model/
    ├── WebhookEvent.java                # Incoming Flodesk event payload
    ├── SubscriberInfo.java              # Subscriber sub-object
    └── WebhookResponse.java             # Outgoing response body
```

## Design Notes

- **Provider Gamma** uses `java.util.Random.nextBoolean()` — 50% failure rate. For a configurable failure rate, pass a `failureRate` environment variable and seed `Random` accordingly.
- The server runs on port **8081** to avoid colliding with the main notifier service on 8080.
- All incoming events are logged at INFO level; failures at WARN level.
- `@JsonIgnoreProperties(ignoreUnknown = true)` on model classes makes the server tolerant of new/unknown fields as the Flodesk schema evolves.
