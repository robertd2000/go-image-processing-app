# =========================================
# LOAD ENV
# =========================================
include .env
export

DC=docker compose

# =========================================
# DATABASE URLs (для контейнеров)
# =========================================
AUTH_DB_URL=postgres://${AUTH_DB_USER}:${AUTH_DB_PASS}@postgres-auth:5432/${AUTH_DB_NAME}?sslmode=disable
IMAGE_DB_URL=postgres://${IMAGE_DB_USER}:${IMAGE_DB_PASS}@postgres-image:5432/${IMAGE_DB_NAME}?sslmode=disable

# =========================================
# DOCKER LIFECYCLE
# =========================================

up:
	$(DC) up -d --build

down:
	$(DC) down

restart:
	$(DC) down
	$(DC) up -d --build

logs:
	$(DC) logs -f

ps:
	$(DC) ps

# =========================================
# MIGRATIONS (через контейнер migrate)
# =========================================

auth-migrate-up:
	docker exec migrate migrate -path=/migrations/auth -database "$(AUTH_DB_URL)" up

auth-migrate-down:
	docker exec migrate migrate -path=/migrations/auth -database "$(AUTH_DB_URL)" down 1

auth-migrate-create:
	docker exec migrate migrate create -ext sql -dir /migrations/auth -seq $(name)

image-migrate-up:
	docker exec migrate migrate -path=/migrations/image -database "$(IMAGE_DB_URL)" up

image-migrate-down:
	docker exec migrate migrate -path=/migrations/image -database "$(IMAGE_DB_URL)" down 1

image-migrate-create:
	docker exec migrate migrate create -ext sql -dir /migrations/image -seq $(name)

# =========================================
# KAFKA TOPICS
# =========================================

topics:
	docker exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server ${KAFKA_BROKER} \
		--replication-factor 1 \
		--partitions 3 \
		--topic ${KAFKA_TOPIC_PROCESS}

	docker exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server ${KAFKA_BROKER} \
		--replication-factor 1 \
		--partitions 3 \
		--topic ${KAFKA_TOPIC_DONE}

# =========================================
# BOOTSTRAP (поднять всё + миграции + топики)
# =========================================

bootstrap:
	make up
	@echo "Waiting 5s for Postgres and Kafka..."
	sleep 5
	make auth-migrate-up
	make image-migrate-up
	make topics

# =========================================
# RESET DATABASES
# =========================================

reset-auth-db:
	docker exec postgres-auth psql -U ${AUTH_DB_USER} -d ${AUTH_DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset-image-db:
	docker exec postgres-image psql -U ${IMAGE_DB_USER} -d ${IMAGE_DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset:
	make reset-auth-db
	make reset-image-db
	make auth-migrate-up
	make image-migrate-up

# =========================================
# HEALTH CHECK
# =========================================

health:
	docker ps

# =========================================
# BUILD SERVICES (опционально)
# =========================================

build-auth:
	cd auth && go build -o bin/auth cmd/main.go

build-image:
	cd image && go build -o bin/image cmd/main.go

build-processor:
	cd processor && go build -o bin/processor cmd/main.go

build: build-auth build-image build-processor
