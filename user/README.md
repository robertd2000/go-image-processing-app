# User Service

User service is responsible for **user profiles, settings, and public user data**.
It also **consumes events from Kafka** and keeps its own user projection in sync.

---

# 🧱 Responsibilities

- User profile management
- User settings management
- Public user data (username, avatar, etc.)
- Consume user events from Kafka (event-driven sync)
- Provide secured API for user data mutation

---

# 🧠 Domain (DDD-lite)

## Aggregate

```
User
 ├── Profile
 └── Settings
```

---

## Key Concepts

- All fields are private
- State changes only via domain methods
- Value Objects enforce validation
- Domain is isolated from infrastructure (DB, Kafka, HTTP)

---

## Value Objects

- Username (3–30 chars)
- Email (validated via `net/mail`)
- UserStatus (`active` / `inactive` / `banned`)

---

# 🗄 Database

Tables:

- `users`
- `user_profiles`
- `user_settings`

---

# 🔄 Event-Driven Integration (Kafka)

User service **does NOT create users directly**.
It listens to events from Auth service.

---

## Consumed Events

### `user.created.v1`

```json
{
  "event_id": "uuid",
  "event_type": "user.created.v1",
  "version": 1,
  "occurred_at": "timestamp",
  "payload": {
    "id": "uuid",
    "username": "string",
    "email": "string",
    "created_at": "timestamp"
  }
}
```

---

## Flow

```
auth-service → Kafka → user-service
```

---

## DLQ (Dead Letter Queue)

If message processing fails after retries:

```
user.created.v1 → ❌ → user.created.dlq
```

Used for:

- debugging
- replaying failed events
- avoiding message loss

---

## Retry Strategy

- 3 retries
- exponential backoff (simple incremental delay)
- then → DLQ

---

# 🔐 Authentication & Authorization

User service uses **JWT validation only**.

- Tokens are issued by **auth-service**
- User service **does NOT generate tokens**

---

## Protected Routes

Require:

```
Authorization: Bearer <access_token>
```

---

## Access Rules

- User can modify only **own data**
- Future: role-based access (admin)

---

# 🔌 API

## Public

```
GET /api/v1/users/:id
GET /api/v1/users
```

---

## Protected

```
PUT /api/v1/users/:id
PUT /api/v1/users/:id/profile
PUT /api/v1/users/:id/settings
DELETE /api/v1/users/:id
```

---

# 📘 Swagger

Generate docs:

```bash
swag init -g cmd/main.go -o docs
```

---

## Authorization in Swagger

Click **Authorize** and enter:

```
Bearer <your_token>
```

---

# ⚙️ Configuration

## ENV (.env)

```env
# =====================
# POSTGRES
# =====================
POSTGRES_HOST=postgres-user
POSTGRES_PORT=5432
POSTGRES_USER=user_user
POSTGRES_PASSWORD=user_pass
POSTGRES_DB_NAME=user_db
POSTGRES_SSL_MODE=disable

# =====================
# SERVER
# =====================
SERVER_PORT=8083
RUN_MODE=debug

# =====================
# KAFKA
# =====================
KAFKA_BROKERS=kafka:9092
KAFKA_GROUP_ID=user-service
KAFKA_TOPICS_USER_CREATED=user.created.v1

# =====================
# JWT
# =====================
JWT_SECRET=super_secret_jwt_key_change_me

# =====================
# LOG
# =====================
LOG_LEVEL=debug
```

---

## YAML config (example)

```yaml
server:
  port: 8083
  run_mode: debug

postgres:
  host: postgres-user
  port: 5432
  user: user_user
  password: user_pass
  db_name: user_db
  ssl_mode: disable

kafka:
  brokers:
    - kafka:9092
  group_id: user-service
  topics:
    user_created: user.created.v1

jwt:
  secret: super_secret_jwt_key_change_me

log:
  level: debug
```

---

# 🚀 Run

### локально

```bash
make up
```

---

### через root проекта

```bash
MSYS_NO_PATHCONV=1 make bootstrap
```

---

# 🧪 Testing

```bash
go test ./internal/domain/...
```

Covers:

- value objects
- aggregate logic
- domain invariants

---

# 🧩 Architecture Notes

- User service is **event-driven (eventual consistency)**
- Auth service is **source of truth for identity**
- User service stores **denormalized user data**
- No direct sync calls between services (only events)

---

# ⚠️ Important

- `JWT_SECRET` must match **auth-service**
- Only **access tokens** are accepted
- Kafka must be available before service starts

---

# 🚀 Next Steps

- Add more events (user.updated, user.deleted)
- Implement event router (multi-event support)
- Add RBAC (roles & permissions)
- Add observability (metrics, tracing)
- Add DLQ reprocessing tool
