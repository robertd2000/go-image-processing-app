# Image Processing Microservices

A microservices-based system for image processing built with Go.

The project demonstrates:

- service isolation (separate DB per service)
- event-driven architecture (Kafka)
- DDD-lite domain modeling (user service)
- Docker-based infrastructure

---

# 🧱 Architecture

```text
auth        → authentication & JWT
user        → user profiles & settings
image       → image metadata & storage
processor   → async image processing (Kafka)
```

---

# 📦 Services

## Auth Service

- registration & login
- JWT (access + refresh)
- roles & permissions

## User Service

- user profile (username, avatar, etc.)
- user settings
- DDD-style domain (aggregate + value objects)

## Image Service

- image metadata
- integration with storage (planned)

## Processor Service

- Kafka consumer
- image processing (resize/compress)

---

# 🧠 Key Concepts

- **Single `user_id` across all services**
- **Auth owns identity**
- **User owns profile**
- **Event-driven communication (planned via Kafka)**

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
# AUTH DB
AUTH_DB_USER=auth_user
AUTH_DB_PASS=auth_pass
AUTH_DB_NAME=auth_db

# USER DB
USER_DB_USER=user_user
USER_DB_PASS=user_pass
USER_DB_NAME=user_db

# IMAGE DB
IMAGE_DB_USER=image_user
IMAGE_DB_PASS=image_pass
IMAGE_DB_NAME=image_db

# JWT
JWT_SECRET=super_secret_jwt_key_change_me

# KAFKA
KAFKA_BROKER=kafka:9092
KAFKA_BROKER_HOST=localhost:29092

KAFKA_TOPIC_PROCESS=image.process
KAFKA_TOPIC_DONE=image.done

# SERVER (default fallback)
SERVER_PORT=8080
```

---

## 3. Service Config

Each service uses:

- `config.yaml` (defaults)
- `.env` (override via Viper)

Example (`user/config.yaml`):

```yaml
server:
  port: 8080
  run_mode: debug

postgres:
  host: postgres-user
  port: 5432
  user: user_user
  password: user_pass
  db_name: user_db
  ssl_mode: disable
```

> ⚠️ No `${VAR}` in YAML — ENV overrides are handled by Viper.

---

## 4. Run Everything

```bash
make bootstrap
```

### On Windows (Git Bash)

```bash
MSYS_NO_PATHCONV=1 make bootstrap
```

---

## 🛠 Useful Commands

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

Run migrations per service:

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

# 📡 Kafka

Topics:

```text
image.process
image.done
```

Planned:

```text
user.created
image.uploaded
```

---

# 📚 API

## Auth

```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
```

Swagger:

```
http://localhost:8080/swagger/index.html
```

---

## User (in progress)

```
GET  /api/users/me
PUT  /api/users/profile
PUT  /api/users/settings
```

---

# 🧪 Testing

Run domain tests:

```bash
go test ./internal/domain/...
```

Covers:

- value objects (username, email)
- user aggregate logic
- domain invariants

---

# 🏗 Services Status

## ✅ Done

- Auth service (JWT, refresh tokens)
- Docker infrastructure
- Migrations setup
- User domain (DDD-lite + tests)

## 🚧 In Progress

- User service (repository, handlers)
- Image service

## 📌 Planned

- Kafka integration (auth → user)
- Image processing pipeline
- Storage (S3 / MinIO)

---

# 🧩 Notes

- Each service has its own PostgreSQL database
- Domain layer is isolated from DB schema
- Config = YAML (defaults) + ENV override
- Start simple → evolve to event-driven

---

# 🚀 Roadmap

- [ ] User repository (Postgres mapping)
- [ ] User HTTP API
- [ ] Kafka integration
- [ ] Image upload & storage
- [ ] Processor workers
- [ ] Monitoring & logging

---

# 💡 Philosophy

This project focuses on:

- real backend architecture (not just CRUD)
- clean separation of concerns
- practical DDD (without overengineering)
- incremental system design

---
