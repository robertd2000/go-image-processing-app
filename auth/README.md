# Auth Service

Authentication service responsible for user identity, JWT issuance, and access control.

---

# 🧱 Responsibilities

- User registration & login
- JWT (access & refresh tokens)
- Token validation & refresh
- Role management

---

# 🧠 Domain

Auth service owns:

```text
identity
 ├── email
 ├── password (hashed)
 ├── roles
 └── tokens
```

---

# 🗄 Database

Main tables:

- `auth_users`
- `roles`
- `user_roles`
- `refresh_tokens`

---

# ⚙️ Configuration

## Environment variables

```env
POSTGRES_HOST=postgres-auth
POSTGRES_PORT=5432
POSTGRES_USER=auth_user
POSTGRES_PASSWORD=auth_pass
POSTGRES_DB_NAME=auth_db
POSTGRES_SSL_MODE=disable

JWT_SECRET=super_secret_jwt_key_change_me
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_MIN=10080

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

# 📡 API

## Register

```http
POST /api/auth/register
```

## Login

```http
POST /api/auth/login
```

## Refresh

```http
POST /api/auth/refresh
```

---

# 🔐 JWT

Access token:

```json
{
  "sub": "user_id",
  "role": "user"
}
```

---

# 🔄 Integration

## With User Service

Current approach:

- HTTP call after registration (planned)

Future:

- Kafka event:

```json
{
  "event": "user.created",
  "user_id": "uuid"
}
```

---

# 🧪 Testing

```bash
go test ./...
```

---

# 🧩 Notes

- Auth is the **source of truth for identity**
- Passwords are stored hashed
- Refresh tokens are persisted
- Access tokens are stateless

---
