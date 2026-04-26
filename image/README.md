# Image Service

Сервис управления изображениями: загрузка, хранение метаданных и инициирование асинхронной обработки изображений.

Сервис является частью микросервисной архитектуры и взаимодействует с другими сервисами через Kafka. Хранение файлов осуществляется в S3-совместимом хранилище (AWS S3 / MinIO / Cloudflare R2).

---

## 📌 Основные возможности

- Загрузка изображений
- Хранение метаданных изображений
- Получение списка изображений пользователя
- Инициирование трансформаций изображений
- Асинхронная обработка через Kafka
- Идемпотентные трансформации (через hash)

---

## 🧱 Архитектура

Сервис построен по принципам:

- Clean Architecture
- DDD (Domain-Driven Design)
- Event-Driven Architecture

### Границы сервиса

Image Service отвечает за:

- управление изображениями (metadata)
- генерацию storage key
- публикацию событий в Kafka

Сервис **не выполняет обработку изображений** — это делает отдельный Processor Service.

---

## 🧩 Взаимодействие сервисов

```text
Client → Image Service → S3 (original upload)
                       → Kafka (event)
Processor → S3 (read original)
          → transform
          → S3 (write result)
          → Kafka (event)
Image Service ← Kafka (update status)
```

---

## 🗂 Структура проекта

```text
internal/
  domain/
    image/            # агрегат Image
    transformation/   # агрегат Transformation
    valueobject/      # VO (StorageKey, TransformSpec и др.)
    events/           # доменные события

  usecase/            # бизнес-сценарии (application layer)
  delivery/           # HTTP handlers / transport
  repository/         # реализация доступа к БД
  storage/            # работа с S3
  kafka/              # продюсеры/консьюмеры
```

---

## 🧠 Доменные модели

### Image (Aggregate)

Описывает загруженное изображение:

- `id`
- `user_id`
- `storage_key`
- `metadata (width, height, size, mime)`
- `created_at`

---

### Transformation (Aggregate)

Описывает задачу на обработку:

- `id`
- `image_id`
- `transform_spec (JSON)`
- `transform_hash`
- `status (pending / processing / done / failed)`
- `result_key`

---

## 💾 Хранилище

Используется S3-совместимое хранилище.

Структура ключей:

```text
originals/{user_id}/{image_id}.{ext}
processed/{image_id}/{transform_hash}.{ext}
```

---

## 🗄 База данных

PostgreSQL

### Основные таблицы:

- `images`
- `transformations`

Особенности:

- UUID в качестве primary key
- JSONB для хранения transform_spec
- уникальный индекс `(image_id, transform_hash)` для идемпотентности

---

## 🔄 Kafka события

### Outgoing

- `transformation.requested`

```json
{
  "transformation_id": "uuid",
  "image_id": "uuid",
  "transform_spec": {}
}
```

---

### Incoming

- `transformation.completed`
- `transformation.failed`

---

## 🌐 API

### Upload image

```
POST /images
Content-Type: multipart/form-data
```

Response:

```json
{
  "id": "uuid",
  "storage_key": "..."
}
```

---

### List images

```
GET /images?page=1&limit=10
```

---

### Get image

```
GET /images/:id
```

---

### Request transformation

```
POST /images/:id/transform
```

```json
{
  "transformations": {
    "resize": { "width": 100, "height": 100 },
    "format": "jpeg"
  }
}
```

---

### Get transformation status

```
GET /transformations/:id
```

---

## ⚙️ Конфигурация

Пример `.env`:

```
APP_PORT=8080

DB_DSN=postgres://user:pass@localhost:5432/images

S3_ENDPOINT=http://localhost:9000
S3_BUCKET=images
S3_ACCESS_KEY=admin
S3_SECRET_KEY=admin

KAFKA_BROKERS=localhost:9092
```

---

## 🚀 Запуск

### Локально

```bash
go run ./cmd/app
```

---

### Через Docker

```bash
docker-compose up --build
```

---

## 🧪 Тестирование

```bash
go test ./...
```

---

## 📈 Production заметки

- Использовать presigned URL для загрузки (избегать проксирования файлов через backend)
- Использовать CDN для отдачи изображений
- Добавить rate limiting на трансформации
- Реализовать retry + DLQ для Kafka
- Обеспечить идемпотентность обработчиков

---

## 🧭 Roadmap

- [ ] Presigned upload
- [ ] CDN integration
- [ ] Webhook / SSE уведомления
- [ ] Batch transformations
- [ ] Image versioning

---

## 📄 License

MIT
