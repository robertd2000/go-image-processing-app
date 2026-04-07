# Auth Service

Authentication service responsible for **user identity, JWT issuance, and access control**.
Acts as the **source of truth for authentication and identity**.

---

# 🧱 Responsibilities

- User registration & login
- JWT (access & refresh tokens)
- Token validation & refresh
- Role management
- Publishing domain events to Kafka (event-driven architecture)

---

# 🧠 Domain

Auth service owns:

```
identity
 ├── email
 ├── password (hashed)
 ├── roles
 └── tokens
```

---

## Key Concepts

- Identity is fully managed inside auth-service
- Passwords are always hashed (never stored in plain text)
- Access tokens are stateless
- Refresh tokens are persisted and validated
- Events are emitted via **outbox pattern**

---

# 🗄 Database

Main tables:

- `auth_users`
- `roles`
- `user_roles`
- `refresh_tokens`
- `outbox_events` ← **важно (для Kafka)**

---

# 🔄 Event-Driven Integration (Kafka)

Auth service publishes events to Kafka.

---

## Produced Events

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
auth-service → outbox_events → Kafka → user-service
```

---

## Outbox Pattern

Used to guarantee **reliable event delivery**.

Flow:

1. User is created in DB
2. Event is saved in `outbox_events`
3. Background worker publishes to Kafka
4. Event is marked as sent

---

## Why Outbox?

- avoids lost events
- ensures consistency between DB and Kafka
- safe retries

---

# 🔐 JWT

Auth service issues two types of tokens:

---

## Access Token

- Short-lived (e.g. 15–30 min)
- Used for API authorization
- Stateless

Example payload:

```json
{
  "user_id": "uuid",
  "jti": "access",
  "exp": 1234567890
}
```

---

## Refresh Token

- Long-lived (e.g. 7 days)
- Stored in DB
- Used to generate new access tokens

---

## Token Rules

- Only access tokens are accepted by other services
- Token type is validated via `jti` field
- Tokens are signed with shared secret (HS256)

---

# 📡 API

## Register

```http
POST /api/v1/auth/register
```

Creates user + emits `user.created.v1`

---

## Login

```http
POST /api/v1/auth/login
```

Returns:

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

---

## Refresh

```http
POST /api/v1/auth/refresh
```

- validates refresh token
- returns new access token

---

# 📘 Swagger

Generate docs:

```bash
swag init -g cmd/main.go -o docs
```

---

# ⚙️ Configuration

## ENV (.env)

```env
# =====================
# POSTGRES
# =====================
POSTGRES_HOST=postgres-auth
POSTGRES_PORT=5432
POSTGRES_USER=auth_user
POSTGRES_PASSWORD=auth_pass
POSTGRES_DB_NAME=auth_db
POSTGRES_SSL_MODE=disable

# =====================
# JWT
# =====================
JWT_SECRET=super_secret_jwt_key_change_me
JWT_ACCESS_TTL_MIN=30
JWT_REFRESH_TTL_MIN=10080

# =====================
# SERVER
# =====================
SERVER_PORT=8080

# =====================
# KAFKA
# =====================
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_USER_CREATED=user.created.v1

# =====================
# LOG
# =====================
LOG_LEVEL=debug
```

---

## YAML config (example)

```yaml
server:
  port: 8080
  run_mode: debug

postgres:
  host: postgres-auth
  port: 5432
  user: auth_user
  password: auth_pass
  db_name: auth_db
  ssl_mode: disable

kafka:
  brokers:
    - kafka:9092
  topics:
    user_created: user.created.v1

jwt:
  secret: super_secret_jwt_key_change_me
  access_ttl: 30m
  refresh_ttl: 168h

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
go test ./...
```

---

# 🧩 Architecture Notes

- Auth service is **source of truth for identity**
- Uses **JWT for stateless authentication**
- Uses **refresh tokens for session continuity**
- Uses **outbox pattern for reliable Kafka publishing**
- Other services must **validate JWT locally** (no HTTP calls)

---

# ⚠️ Important

- `JWT_SECRET` must match all services (user, image, etc.)
- Kafka must be available for event publishing
- Refresh tokens must be securely stored
- Access tokens must not be persisted

---

# 🚀 Next Steps

- Add more events (user.updated, user.deleted)
- Implement token revocation / blacklist
- Add device/session tracking
- Add rate limiting for auth endpoints
- Add audit logging
