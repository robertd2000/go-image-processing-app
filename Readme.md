# Image Processing Microservices

This project is a microservices-based system that includes:

- **Auth service** (authentication, JWT)
- **Image service** (image storage & metadata)
- **Processor service** (async image processing via Kafka)
- **PostgreSQL** (separate DBs per service)
- **Kafka** (event-driven processing)

---

## 🚀 Getting Started

### 1. Requirements

- Docker & Docker Compose
- Make

---

### 2. Setup environment

Create a `.env` file in the root (or use the existing one):

```env
# AUTH DB
AUTH_DB_USER=auth_user
AUTH_DB_PASS=auth_pass
AUTH_DB_NAME=auth_db

# IMAGE DB
IMAGE_DB_USER=image_user
IMAGE_DB_PASS=image_pass
IMAGE_DB_NAME=image_db

# JWT
JWT_SECRET=super_secret_jwt_key_change_me

# KAFKA
KAFKA_BROKER=kafka:9092

# SERVER
SERVER_PORT=8080
```

---

### 3. Run everything

```bash
make bootstrap
```

This will:

- build and start all containers
- wait for databases
- run migrations
- create Kafka topics

---

## 🪟 Windows Note (Git Bash)

If you are using **Git Bash on Windows**, use:

```bash
MSYS_NO_PATHCONV=1 make bootstrap
```

Or export once:

```bash
export MSYS_NO_PATHCONV=1
make bootstrap
```

---

## 🛠 Useful Commands

### Start / Stop

```bash
make up
make down
make restart
```

---

### Logs

```bash
make logs
```

---

### Check running containers

```bash
make ps
```

---

### Health check

```bash
make health
```

---

## 🗄 Migrations

### Run migrations

```bash
make auth-migrate-up
make image-migrate-up
```

### Rollback

```bash
make auth-migrate-down
make image-migrate-down
```

### Create migration

```bash
make auth-migrate-create name=your_migration_name
make image-migrate-create name=your_migration_name
```

---

## 🔄 Reset Databases

```bash
make reset
```

This will:

- drop schemas
- re-run migrations

---

## 📡 Kafka Topics

```bash
make topics
```

Topics used:

- `image.process`
- `image.done`

---

## 📚 API Documentation

Swagger is available at:

```
http://localhost:8080/swagger/index.html
```

---

## 🏗 Services

- **Auth Service** → authentication & JWT
- **Image Service** → images and metadata
- **Processor Service** → async processing

---

## ✅ Services Status

- [x] **Auth Service**
  - JWT authentication
  - Access / Refresh tokens
  - User management (basic)

- [ ] **Image Service**
  - Upload images
  - Store metadata
  - API for image retrieval

- [ ] **Processor Service**
  - Consume Kafka events
  - Process images (resize/compress)
  - Publish results

---

## 🧭 Roadmap (optional)

- [ ] Add image storage (S3 / local FS)
- [ ] Improve error handling & retries
- [ ] Add monitoring (Prometheus / Grafana)
- [ ] Add tracing (Jaeger)
- [ ] Add rate limiting

---

## ⚙️ Build Services (optional)

```bash
make build
```

Or individually:

```bash
make build-auth
make build-image
make build-processor
```

---

## 🧩 Notes

- Each service has its own PostgreSQL database
- Communication between services is done via Kafka
- JWT is used for authentication
