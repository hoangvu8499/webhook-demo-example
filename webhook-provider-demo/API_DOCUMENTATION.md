# API Documentation — webhook-provider-demo

## Mục lục

1. [Tổng quan](#1-tổng-quan)
2. [Thông tin chung](#2-thông-tin-chung)
3. [Schema dữ liệu](#3-schema-dữ-liệu)
   - [Request Body — WebhookEvent](#31-request-body--webhookevent)
   - [Response Body — WebhookResponse](#32-response-body--webhookresponse)
4. [Danh sách Endpoints](#4-danh-sách-endpoints)
   - [POST /webhook/provider-alpha](#41-post-webhookprovider-alpha)
   - [POST /webhook/provider-beta](#42-post-webhookprovider-beta)
   - [POST /webhook/provider-gamma](#43-post-webhookprovider-gamma)
5. [HTTP Status Codes](#5-http-status-codes)
6. [Ví dụ Request theo loại Event](#6-ví-dụ-request-theo-loại-event)
7. [Logging](#7-logging)

---

## 1. Tổng quan

**webhook-provider-demo** là một mock server mô phỏng **3 endpoint nhận webhook** từ 3 nhà cung cấp khác nhau. Server này được thiết kế để hỗ trợ kiểm thử hệ thống **Webhook Notifier** — đặc biệt là các cơ chế retry, backoff, và fairness khi gửi webhook đến các provider bên ngoài.

Ba endpoint hoạt động với các hành vi khác nhau:

| Provider | Endpoint | Hành vi |
|---|---|---|
| **Provider Alpha** | `POST /webhook/provider-alpha` | Luôn trả về **SUCCESS** (HTTP 200) |
| **Provider Beta** | `POST /webhook/provider-beta` | Luôn trả về **FAIL** (HTTP 500) |
| **Provider Gamma** | `POST /webhook/provider-gamma` | Ngẫu nhiên **SUCCESS hoặc FAIL** (~50/50) |

> **Mục đích sử dụng:**
> - **Provider Alpha** → kiểm tra luồng bình thường (happy path)
> - **Provider Beta** → kiểm tra cơ chế retry khi endpoint luôn lỗi
> - **Provider Gamma** → kiểm tra retry + exponential backoff với endpoint không ổn định

---

## 2. Thông tin chung

| Thuộc tính | Giá trị |
|---|---|
| Base URL | `http://localhost:8081` |
| Protocol | HTTP |
| Content-Type | `application/json` |
| Method | `POST` (tất cả endpoints) |
| Java Version | 8 |
| Framework | Spring Boot 2.7.18 (Spring Framework 5.3.x) |

**Header bắt buộc:**

```
Content-Type: application/json
```

---

## 3. Schema dữ liệu

### 3.1 Request Body — WebhookEvent

Tất cả 3 endpoint nhận cùng một cấu trúc payload. Schema khớp với định dạng webhook event của Flodesk (theo `openapi.json`).

```json
{
  "event_name": "string",
  "event_time": "string (ISO 8601)",
  "webhook_id": "string",
  "subscriber": { ... },
  "segment": { ... }
}
```

#### Bảng mô tả các field

| Field | Kiểu | Bắt buộc | Mô tả |
|---|---|---|---|
| `event_name` | `string` | Có | Tên sự kiện. Các giá trị hợp lệ: `subscriber.created`, `subscriber.unsubscribed`, `subscriber.added_to_segment` |
| `event_time` | `string` | Có | Thời điểm xảy ra sự kiện, định dạng ISO 8601. Ví dụ: `2024-01-02T15:04:05.999Z` |
| `webhook_id` | `string` | Có | ID của webhook đã kích hoạt event này |
| `subscriber` | `object` | Có | Thông tin subscriber liên quan đến event (xem chi tiết bên dưới) |
| `segment` | `object` | Không | Chỉ có trong event `subscriber.added_to_segment`. Thông tin segment mà subscriber vừa được thêm vào |

#### Object: `subscriber`

| Field | Kiểu | Bắt buộc | Mô tả |
|---|---|---|---|
| `id` | `string` | Có | ID duy nhất của subscriber |
| `email` | `string` | Có | Địa chỉ email |
| `status` | `string` | Không | Trạng thái. Giá trị: `active`, `unsubscribed`, `unconfirmed`, `bounced`, `complained`, `cleaned` |
| `source` | `string` | Không | Nguồn gốc subscriber. Giá trị: `manual`, `csv`, `form_optin`, `integration`, `checkout` |
| `first_name` | `string` | Không | Tên |
| `last_name` | `string` | Không | Họ |
| `segments` | `array` | Không | Danh sách segment mà subscriber thuộc về |
| `custom_fields` | `object` | Không | Các trường tuỳ chỉnh dạng key-value (`string: string`) |
| `optin_ip` | `string` | Không | Địa chỉ IP khi subscriber xác nhận opt-in |
| `optin_timestamp` | `string` | Không | Thời điểm xác nhận opt-in, định dạng ISO 8601 |
| `created_at` | `string` | Không | Thời điểm tạo subscriber, định dạng ISO 8601 |

#### Object: `segment` (chỉ có trong `subscriber.added_to_segment`)

| Field | Kiểu | Mô tả |
|---|---|---|
| `id` | `string` | ID của segment |
| `name` | `string` | Tên segment |

---

### 3.2 Response Body — WebhookResponse

Tất cả 3 endpoint đều trả về cùng cấu trúc response.

```json
{
  "status": "SUCCESS | FAIL",
  "message": "string",
  "provider": "string",
  "received_event": "string",
  "received_at": "string (ISO 8601)"
}
```

| Field | Kiểu | Mô tả |
|---|---|---|
| `status` | `string` | Kết quả xử lý. Giá trị: `SUCCESS` hoặc `FAIL` |
| `message` | `string` | Mô tả kết quả |
| `provider` | `string` | Tên provider nhận event. Giá trị: `provider-alpha`, `provider-beta`, `provider-gamma` |
| `received_event` | `string` | Tên event đã nhận được (lấy từ `event_name` của request) |
| `received_at` | `string` | Thời điểm server nhận request, định dạng ISO 8601 |

---

## 4. Danh sách Endpoints

### 4.1 POST /webhook/provider-alpha

**Mô tả:** Mô phỏng nhà cung cấp ổn định, luôn xử lý thành công mọi webhook nhận được.

**Hành vi:** Luôn trả về `HTTP 200 OK` với `status: SUCCESS`.

**Use case:** Kiểm tra luồng bình thường (happy path) — xác nhận notifier gửi đúng payload và xử lý response 2XX đúng cách.

---

**Request**

```
POST http://localhost:8081/webhook/provider-alpha
Content-Type: application/json
```

**Request Body:**

```json
{
  "event_name": "subscriber.created",
  "event_time": "2024-06-29T10:00:00.000Z",
  "webhook_id": "wh_abc123",
  "subscriber": {
    "id": "sub_001",
    "email": "john.doe@example.com",
    "status": "active",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-06-29T10:00:00.000Z"
  }
}
```

**Response — HTTP 200 OK**

```json
{
  "status": "SUCCESS",
  "message": "Event received and processed successfully.",
  "provider": "provider-alpha",
  "received_event": "subscriber.created",
  "received_at": "2024-06-29T10:00:00.123456789Z"
}
```

---

### 4.2 POST /webhook/provider-beta

**Mô tả:** Mô phỏng nhà cung cấp lỗi hoàn toàn, không thể xử lý bất kỳ webhook nào.

**Hành vi:** Luôn trả về `HTTP 500 Internal Server Error` với `status: FAIL`.

**Use case:** Kiểm tra cơ chế retry — notifier phải nhận diện được 5XX và thực hiện retry với exponential backoff, cuối cùng đánh dấu `failed_permanently` sau khi hết số lần thử.

---

**Request**

```
POST http://localhost:8081/webhook/provider-beta
Content-Type: application/json
```

**Request Body:**

```json
{
  "event_name": "subscriber.unsubscribed",
  "event_time": "2024-06-29T10:05:00.000Z",
  "webhook_id": "wh_abc456",
  "subscriber": {
    "id": "sub_002",
    "email": "jane.smith@example.com",
    "status": "unsubscribed"
  }
}
```

**Response — HTTP 500 Internal Server Error**

```json
{
  "status": "FAIL",
  "message": "Internal server error: provider endpoint is unavailable.",
  "provider": "provider-beta",
  "received_event": "subscriber.unsubscribed",
  "received_at": "2024-06-29T10:05:00.234567890Z"
}
```

---

### 4.3 POST /webhook/provider-gamma

**Mô tả:** Mô phỏng nhà cung cấp không ổn định, ngẫu nhiên thành công hoặc thất bại.

**Hành vi:** Mỗi request có xác suất ~50% trả về `HTTP 200 OK` (SUCCESS) và ~50% trả về `HTTP 500` (FAIL). Kết quả được quyết định bởi `Random.nextBoolean()`.

**Use case:** Kiểm tra retry + backoff trong điều kiện thực tế — notifier phải retry khi nhận 5XX và dừng khi nhận được 2XX.

---

**Request**

```
POST http://localhost:8081/webhook/provider-gamma
Content-Type: application/json
```

**Request Body:**

```json
{
  "event_name": "subscriber.added_to_segment",
  "event_time": "2024-06-29T10:10:00.000Z",
  "webhook_id": "wh_abc789",
  "subscriber": {
    "id": "sub_003",
    "email": "alice@example.com",
    "status": "active",
    "first_name": "Alice"
  },
  "segment": {
    "id": "seg_001",
    "name": "VIP Customers"
  }
}
```

**Response — HTTP 200 OK (khi thành công)**

```json
{
  "status": "SUCCESS",
  "message": "Event received and processed successfully.",
  "provider": "provider-gamma",
  "received_event": "subscriber.added_to_segment",
  "received_at": "2024-06-29T10:10:00.345678901Z"
}
```

**Response — HTTP 500 Internal Server Error (khi thất bại)**

```json
{
  "status": "FAIL",
  "message": "Transient error: provider failed to process the event. Please retry.",
  "provider": "provider-gamma",
  "received_event": "subscriber.added_to_segment",
  "received_at": "2024-06-29T10:10:00.345678901Z"
}
```

---

## 5. HTTP Status Codes

| HTTP Status | Ý nghĩa | Provider trả về |
|---|---|---|
| `200 OK` | Xử lý thành công | Alpha (luôn), Gamma (ngẫu nhiên) |
| `500 Internal Server Error` | Xử lý thất bại | Beta (luôn), Gamma (ngẫu nhiên) |
| `400 Bad Request` | Request body không hợp lệ hoặc thiếu `Content-Type: application/json` | Cả 3 (Spring tự xử lý) |

> **Lưu ý cho Webhook Notifier:** Chỉ các status code trong dải **2XX** mới được coi là thành công. Mọi response 4XX hoặc 5XX đều cần được đưa vào retry queue.

---

## 6. Ví dụ Request theo loại Event

### Event: `subscriber.created`

```bash
curl -s -X POST http://localhost:8081/webhook/provider-alpha \
  -H "Content-Type: application/json" \
  -d '{
    "event_name": "subscriber.created",
    "event_time": "2024-06-29T10:00:00.000Z",
    "webhook_id": "wh_001",
    "subscriber": {
      "id": "sub_001",
      "email": "newuser@example.com",
      "status": "active",
      "source": "form_optin",
      "first_name": "New",
      "last_name": "User",
      "custom_fields": {
        "favorite_color": "Blue"
      },
      "optin_ip": "192.168.1.1",
      "optin_timestamp": "2024-06-29T09:59:00.000Z",
      "created_at": "2024-06-29T10:00:00.000Z"
    }
  }'
```

---

### Event: `subscriber.unsubscribed`

```bash
curl -s -X POST http://localhost:8081/webhook/provider-beta \
  -H "Content-Type: application/json" \
  -d '{
    "event_name": "subscriber.unsubscribed",
    "event_time": "2024-06-29T11:00:00.000Z",
    "webhook_id": "wh_002",
    "subscriber": {
      "id": "sub_002",
      "email": "leaving@example.com",
      "status": "unsubscribed"
    }
  }'
```

---

### Event: `subscriber.added_to_segment`

```bash
curl -s -X POST http://localhost:8081/webhook/provider-gamma \
  -H "Content-Type: application/json" \
  -d '{
    "event_name": "subscriber.added_to_segment",
    "event_time": "2024-06-29T12:00:00.000Z",
    "webhook_id": "wh_003",
    "subscriber": {
      "id": "sub_003",
      "email": "vip@example.com",
      "status": "active",
      "first_name": "VIP",
      "last_name": "Member"
    },
    "segment": {
      "id": "seg_vip",
      "name": "VIP Customers"
    }
  }'
```

---

## 7. Logging

Server ghi log mỗi request nhận được ra console theo định dạng:

```
yyyy-MM-dd HH:mm:ss [thread] LEVEL logger - message
```

**Ví dụ log thành công (Provider Alpha / Gamma):**

```
2024-06-29 10:00:00 [http-nio-8081-exec-1] INFO  c.d.w.controller.WebhookReceiverController - [provider-alpha] Received event: subscriber.created | webhook_id: wh_001
```

**Ví dụ log thất bại (Provider Beta):**

```
2024-06-29 10:05:00 [http-nio-8081-exec-2] WARN  c.d.w.controller.WebhookReceiverController - [provider-beta] Received event: subscriber.unsubscribed | webhook_id: wh_002 — responding with FAIL
```

Log giúp xác nhận rằng Webhook Notifier đã gửi đúng payload đến đúng endpoint trong quá trình kiểm thử.
