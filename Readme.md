# Image Processing Microservices

A production-style backend project built with Go that demonstrates how a distributed image processing platform can be designed using microservices and event-driven architecture.

The system is designed around independent services, asynchronous communication, and clear ownership of data. Every service owns its own database, communicates through Kafka events, and can evolve independently.

---

# ✨ Features

- Microservices architecture
- Event-driven communication with Kafka
- Database per Service pattern
- Outbox Pattern for reliable event publishing
- JWT authentication
- Refresh tokens
- Role-based authorization foundation
- Dockerized development environment
- PostgreSQL
- MinIO object storage
- Swagger API documentation
- Configuration via YAML + ENV
- Database migrations
- Clean Architecture
- DDD-lite

---

# 🛠 Tech Stack

| Category         | Technology              |
| ---------------- | ----------------------- |
| Language         | Go                      |
| HTTP             | Gin                     |
| Database         | PostgreSQL              |
| Messaging        | Kafka                   |
| Object Storage   | MinIO                   |
| Authentication   | JWT                     |
| Configuration    | Viper                   |
| Documentation    | Swagger                 |
| Migrations       | golang-migrate          |
| Containerization | Docker & Docker Compose |

---

# 🧱 System Architecture

```text
                +----------------+
                |    Auth API    |
                +----------------+
                        |
                 PostgreSQL(Auth)
                        |
                     Outbox
                        |
                        ▼
                    Kafka Topics
                        |
        +---------------+----------------+
        |                                |
        ▼                                ▼
+---------------+               +----------------+
| User Service  |               | Image Service  |
+---------------+               +----------------+
        |                                |
 PostgreSQL(User)              PostgreSQL(Image)
                                         |
                                         ▼
                                  image.process
                                         |
                                         ▼
                               +------------------+
                               | Processor Worker |
                               +------------------+
                                         |
                                  image.processed
                                         |
                                         ▼
                                  Image Service
```

---

# 📦 Services

## Auth Service

Responsible for identity management.

### Features

- User registration
- User login
- JWT generation
- Refresh tokens
- Password hashing
- Role management
- User creation events
- Outbox pattern

Database:

- auth_users
- refresh_tokens
- roles
- user_roles
- outbox

---

## User Service

Maintains a projection of user information.

### Features

- User profile
- User settings
- Avatar
- JWT middleware
- Kafka consumer
- User projection
- Eventual consistency

Database:

- users
- settings

Consumes:

```
user.created.v1
```

---

## Image Service

Responsible for image lifecycle.

### Features

- Upload metadata
- Ownership validation
- Image status
- Object storage integration
- Produce processing events
- Receive processing results

Database:

- images

Produces:

```
image.process
```

Consumes:

```
image.done
```

---

## Processor Service

Background worker for image processing.

### Features

- Kafka consumer
- Resize
- Compression
- Thumbnail generation
- Metadata extraction
- Processing status updates

Consumes:

```
image.process
```

Produces:

```
image.done
```

---

# 🧠 Architecture Principles

- Database per Service
- Event-Driven Architecture
- Eventual Consistency
- Outbox Pattern
- Repository Pattern
- Dependency Injection
- Clean Architecture
- DDD-lite
- Local JWT Validation
- Single Source of Truth for Identity

---

# 🔄 Event Flow

## User Registration

```text
Client
   │
   ▼
Auth Service
   │
   ▼
PostgreSQL
   │
   ▼
Outbox
   │
   ▼
Kafka
   │
   ▼
User Service
   │
   ▼
User Database
```

---

## Image Processing

```text
Client
   │
   ▼
Image Service
   │
   ▼
PostgreSQL
   │
   ▼
Kafka (image.process)
   │
   ▼
Processor
   │
   ▼
MinIO
   │
   ▼
Kafka (image.done)
   │
   ▼
Image Service
```

---

# 📡 Kafka Topics

| Topic           | Description                |
| --------------- | -------------------------- |
| user.created.v1 | User registration          |
| image.process   | Start image processing     |
| image.done      | Image processing completed |

---

# ☠ Dead Letter Queue

Failed messages are redirected into dedicated DLQ topics.

Example:

```text
user.created.v1
        │
        ▼
 Processing Error
        │
        ▼
user.created.dlq
```

---

# 🔐 Authentication

Authentication is centralized in the Auth Service.

Other services never call Auth directly.

Every service validates JWT locally using the shared secret.

Supported tokens:

- Access Token
- Refresh Token

Protected requests:

```http
Authorization: Bearer <access_token>
```

---

# ⚙ Configuration

Configuration is loaded using:

- YAML files
- Environment variables

Priority:

```
ENV > YAML defaults
```

---

Example:

```yaml
server:
  port: 8083

postgres:
  host: postgres-user
  port: 5432

jwt:
  secret: your-secret

kafka:
  brokers:
    - kafka:9092
```

---

# 📁 Repository Structure

```text
.
├── auth/
├── user/
├── image/
├── processor/
├── docker/
├── docs/
├── scripts/
├── Makefile
├── docker-compose.yml
└── .env
```

---

# 🚀 Getting Started

## Requirements

- Go 1.24+
- Docker
- Docker Compose
- Make

---

## Clone

```bash
git clone https://github.com/your-name/image-processing-microservices.git

cd image-processing-microservices
```

---

## Environment

Create root `.env`

```env
AUTH_DB_USER=auth_user
AUTH_DB_PASS=auth_pass
AUTH_DB_NAME=auth_db

USER_DB_USER=user_user
USER_DB_PASS=user_pass
USER_DB_NAME=user_db

IMAGE_DB_USER=image_user
IMAGE_DB_PASS=image_pass
IMAGE_DB_NAME=image_db

JWT_SECRET=super_secret

KAFKA_BROKER=kafka:9092
KAFKA_BROKER_HOST=localhost:29092

KAFKA_TOPIC_USER_CREATED=user.created.v1
KAFKA_TOPIC_PROCESS=image.process
KAFKA_TOPIC_DONE=image.done
```

---

# 🚀 Run

Bootstrap everything

```bash
make bootstrap
```

Windows (Git Bash)

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

make reset
```

---

# 🗄 Database Migrations

Run migrations

```bash
make auth-migrate-up

make user-migrate-up

make image-migrate-up
```

Rollback

```bash
make auth-migrate-down

make user-migrate-down

make image-migrate-down
```

Create migration

```bash
make user-migrate-create name=create_users
```

---

# 📖 API

## Auth

```
POST /api/v1/auth/register

POST /api/v1/auth/login

POST /api/v1/auth/refresh
```

Swagger

```
http://localhost:8080/swagger/index.html
```

---

## User

```
GET /api/v1/users

GET /api/v1/users/{id}

PUT /api/v1/users/{id}

PUT /api/v1/users/{id}/profile

PUT /api/v1/users/{id}/settings

DELETE /api/v1/users/{id}
```

Swagger

```
http://localhost:8083/swagger/index.html
```

---

## Image

Planned endpoints

```
POST /api/v1/images

GET /api/v1/images

GET /api/v1/images/{id}

DELETE /api/v1/images/{id}
```

---

# 🧪 Testing

Run all tests

```bash
go test ./...
```

Current test coverage focuses on

- Domain logic
- Aggregates
- Value Objects
- Repository layer

---

# 📦 Infrastructure

Docker Compose starts:

- Auth PostgreSQL
- User PostgreSQL
- Image PostgreSQL
- Kafka
- Zookeeper
- Kafka UI
- MinIO
- All services

---

# 📈 Future Improvements

- Prometheus metrics
- Grafana dashboards
- OpenTelemetry tracing
- Distributed logging
- CI/CD pipeline
- Kubernetes deployment
- gRPC communication
- Saga orchestration
- Image versioning
- Image deduplication
- RBAC
- Rate limiting

---

# ✅ Project Status

### Completed

- Authentication
- JWT
- Refresh Tokens
- Role Management
- Outbox Pattern
- Kafka Integration
- User Projection
- Docker Infrastructure
- Swagger
- Database Migrations

---

### In Progress

- Image Service
- Processor Worker
- MinIO Integration

---

### Planned

- Thumbnail generation
- Image transformations
- Observability
- Monitoring
- Tracing
- Kubernetes
- CI/CD

---

# 💡 Design Goals

This project is intended to demonstrate production-ready backend architecture rather than simple CRUD applications.

Key goals:

- Build independent services
- Minimize service coupling
- Favor asynchronous communication
- Keep data ownership explicit
- Apply practical DDD
- Use patterns commonly found in production Go systems

---

# 📄 License

MIT
