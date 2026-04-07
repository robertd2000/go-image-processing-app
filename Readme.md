# Image Processing Microservices

A microservices-based system for image processing built with Go.

The project demonstrates:

- service isolation (separate DB per service)
- event-driven architecture (Kafka + outbox pattern)
- DDD-lite domain modeling
- JWT-based authentication (Auth service as identity provider)
- Docker-based infrastructure

---

# 🧱 Architecture

```text
auth        → authentication, JWT, identity (source of truth)
user        → user profiles & settings (projection via Kafka)
image       → image metadata & storage
processor   → async image processing (Kafka consumer)
```

---

# 📦 Services

## Auth Service

- registration & login
- JWT (access + refresh tokens)
- refresh token storage
- role management
- publishes events (`user.created.v1`)
- uses **outbox pattern**

---

## User Service

- user profile (username, avatar, etc.)
- user settings
- consumes Kafka events
- maintains user projection
- protected API via JWT
- DDD-lite (aggregate + value objects)

---

## Image Service

- image metadata
- image ownership (user_id)
- produces events for processing
- integration with storage (planned)

---

## Processor Service

- Kafka consumer
- async image processing (resize/compress/etc.)
- emits processing results

---

# 🧠 Key Concepts

- **Single `user_id` across all services**
- **Auth = source of truth for identity**
- **User = projection (eventual consistency)**
- **Event-driven communication via Kafka**
- **No synchronous coupling between services**
- **JWT is validated locally (no auth calls between services)**

---

# 🔄 Event-Driven Architecture

## Example flow

```text
[REGISTER]
auth → DB → outbox → Kafka → user → DB

[IMAGE UPLOAD - planned]
image → Kafka → processor → Kafka → image
```

---

## Kafka Topics

```text
user.created.v1
image.process
image.done
```

---

## DLQ (Dead Letter Queue)

Used when message processing fails:

```text
user.created.v1 → ❌ → user.created.dlq
```

---

## Guarantees

- at-least-once delivery
- retry + DLQ fallback
- eventual consistency

---

# 🔐 Authentication

- JWT issued by **auth service**
- Other services only **validate tokens locally**
- Shared secret (`JWT_SECRET`) must be identical across services

---

## Token Types

- access token (short-lived)
- refresh token (long-lived, stored in DB)

---

## Protected Requests

```http
Authorization: Bearer <access_token>
```

---

# 🚀 Getting Started

## 1. Requirements

- Docker
- Docker Compose
- Make

---

## 2. Environment

### Root `.env`

```env
# =====================
# AUTH DB
# =====================
AUTH_DB_USER=auth_user
AUTH_DB_PASS=auth_pass
AUTH_DB_NAME=auth_db

# =====================
# USER DB
# =====================
USER_DB_USER=user_user
USER_DB_PASS=user_pass
USER_DB_NAME=user_db

# =====================
# IMAGE DB
# =====================
IMAGE_DB_USER=image_user
IMAGE_DB_PASS=image_pass
IMAGE_DB_NAME=image_db

# =====================
# JWT (shared!)
# =====================
JWT_SECRET=super_secret_jwt_key_change_me

# =====================
# KAFKA
# =====================
KAFKA_BROKER=kafka:9092
KAFKA_BROKER_HOST=localhost:29092

# topics
KAFKA_TOPIC_USER_CREATED=user.created.v1
KAFKA_TOPIC_PROCESS=image.process
KAFKA_TOPIC_DONE=image.done

# =====================
# SERVER (fallback)
# =====================
SERVER_PORT=8080
```

---

# ⚙️ Configuration

Each service uses:

- `config-*.yml` (defaults)
- `.env` (override via Viper)

---

## Example (`user/config-development.yml`)

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

# 🚀 Run Everything

```bash
make bootstrap
```

---

### Windows (Git Bash)

```bash
MSYS_NO_PATHCONV=1 make bootstrap
```

---

# 🛠 Useful Commands

```bash
make up
make down
make restart
make logs
make ps
make health
```

---

# 🗄 Migrations

```bash
make auth-migrate-up
make user-migrate-up
make image-migrate-up
```

Rollback:

```bash
make auth-migrate-down
make user-migrate-down
make image-migrate-down
```

Create migration:

```bash
make user-migrate-create name=create_users
```

---

# 🔄 Reset Databases

```bash
make reset
```

---

# 📡 API

## Auth

```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
```

Swagger:

```
http://localhost:8080/swagger/index.html
```

---

## User

```
GET    /api/v1/users/:id
GET    /api/v1/users

PUT    /api/v1/users/:id
PUT    /api/v1/users/:id/profile
PUT    /api/v1/users/:id/settings
DELETE /api/v1/users/:id
```

Swagger:

```
http://localhost:8083/swagger/index.html
```

---

# 🧪 Testing

```bash
go test ./...
```

Covers:

- domain logic
- value objects
- aggregates

---

# 🏗 Services Status

## ✅ Done

- Auth service (JWT, refresh tokens, outbox)
- Kafka infrastructure
- User service (consumer + API + auth middleware)
- Docker setup
- Migrations

---

## 🚧 In Progress

- Image service
- Processor service

---

## 📌 Planned

- More events (user.updated, image.uploaded)
- Event router (multi-event handling)
- Image storage (S3 / MinIO)
- Observability (metrics, tracing)
- RBAC (roles & permissions)

---

# 🧩 Notes

- Each service has its own PostgreSQL database
- No shared DB between services
- Communication via Kafka (event-driven)
- Config = YAML + ENV override
- System is eventually consistent

---

# 💡 Philosophy

This project focuses on:

- real backend architecture (not CRUD-only)
- clean separation of concerns
- practical DDD (without overengineering)
- event-driven thinking
- incremental system design

---

# 🚀 Roadmap

- [x] Auth (JWT + refresh)
- [x] Kafka integration
- [x] User projection via events
- [ ] Image upload pipeline
- [ ] Async processing
- [ ] Storage integration
- [ ] Monitoring & tracing
