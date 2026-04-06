# User Service

User service responsible for user profiles, settings, and public user data.

---

# 🧱 Responsibilities

- User profile management
- User settings
- Public user data (username, avatar, etc.)

---

# 🧠 Domain (DDD-lite)

## Aggregate

```text
User
 ├── Profile
 └── Settings
```

---

## Key Concepts

- All fields are private
- State changes only via methods
- Value objects enforce validation
- Domain is isolated from DB

---

## Value Objects

- `Username` (3–30 chars)
- `Email` (validated via `net/mail`)
- `UserStatus` (active / inactive / banned)

---

# 🗄 Database

Tables:

- `users`
- `user_profiles`
- `user_settings`

---

# ⚙️ Configuration

```env
POSTGRES_HOST=postgres-user
POSTGRES_PORT=5432
POSTGRES_USER=user_user
POSTGRES_PASSWORD=user_pass
POSTGRES_DB_NAME=user_db
POSTGRES_SSL_MODE=disable

SERVER_PORT=8080
```

---

# 🚀 Run

```bash
make up
```

Or via root:

```bash
make bootstrap
```

---

# 🔌 API (planned)

```http
GET  /api/users/me
PUT  /api/users/profile
PUT  /api/users/settings
```

---

# 🔄 Integration

## With Auth Service

- Uses same `user_id`
- Auth is source of identity
- User stores profile data

Future:

```text
auth → Kafka → user
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

# 🧩 Notes

- User service does NOT handle authentication
- Email is stored as optional cache
- Domain layer is independent from persistence

---

```bash
$ swag init -g cmd/main.go -o docs
```

---

# 🚀 Next Steps

- Implement repository (Postgres mapping)
- Add HTTP handlers
- Integrate with auth (Kafka or HTTP)

---
