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
USER_DB_URL=postgres://${USER_DB_USER}:${USER_DB_PASS}@postgres-user:5432/${USER_DB_NAME}?sslmode=disable
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
# WAITERS
# =========================================

wait-db:
	@echo "Waiting for Postgres..."
	until $(DC) exec postgres-auth pg_isready -U ${AUTH_DB_USER} >/dev/null 2>&1; do sleep 2; done
	until $(DC) exec postgres-user pg_isready -U ${USER_DB_USER} >/dev/null 2>&1; do sleep 2; done
	until $(DC) exec postgres-image pg_isready -U ${IMAGE_DB_USER} >/dev/null 2>&1; do sleep 2; done
	@echo "Postgres is ready"

wait-kafka:
	@echo "Waiting for Kafka..."
	until $(DC) exec kafka bash -c "kafka-topics --bootstrap-server kafka:9092 --list" >/dev/null 2>&1; do \
		sleep 3; \
	done
	@echo "Kafka is ready"

# =========================================
# MIGRATIONS
# =========================================

auth-migrate-up:
	$(DC) exec migrate migrate -path=/migrations/auth -database "$(AUTH_DB_URL)" up

auth-migrate-down:
	$(DC) exec migrate migrate -path=/migrations/auth -database "$(AUTH_DB_URL)" down 1

auth-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/auth -seq $(name)

user-migrate-up:
	$(DC) exec migrate migrate -path=/migrations/user -database "$(USER_DB_URL)" up

user-migrate-down:
	$(DC) exec migrate migrate -path=/migrations/user -database "$(USER_DB_URL)" down 1

user-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/user -seq $(name)

image-migrate-up:
	$(DC) exec migrate migrate -path=/migrations/image -database "$(IMAGE_DB_URL)" up

image-migrate-down:
	$(DC) exec migrate migrate -path=/migrations/image -database "$(IMAGE_DB_URL)" down 1

image-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/image -seq $(name)

# =========================================
# KAFKA TOPICS
# =========================================

topics:
	$(DC) exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server ${KAFKA_BROKER} \
		--replication-factor 1 \
		--partitions 3 \
		--topic ${KAFKA_TOPIC_PROCESS}

	$(DC) exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server ${KAFKA_BROKER} \
		--replication-factor 1 \
		--partitions 3 \
		--topic ${KAFKA_TOPIC_DONE}

# =========================================
# BOOTSTRAP
# =========================================

bootstrap:
	$(MAKE) up
	$(MAKE) wait-db
	$(MAKE) wait-kafka
	$(MAKE) auth-migrate-up
	$(MAKE) user-migrate-up
	$(MAKE) image-migrate-up
	$(MAKE) topics

# =========================================
# RESET DATABASES
# =========================================

reset-auth-db:
	$(DC) exec postgres-auth psql -U ${AUTH_DB_USER} -d ${AUTH_DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset-user-db:
	$(DC) exec postgres-user psql -U ${USER_DB_USER} -d ${USER_DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset-image-db:
	$(DC) exec postgres-image psql -U ${IMAGE_DB_USER} -d ${IMAGE_DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset:
	$(MAKE) reset-auth-db
	$(MAKE) reset-user-db
	$(MAKE) reset-image-db
	$(MAKE) auth-migrate-up
	$(MAKE) user-migrate-up
	$(MAKE) image-migrate-up

# =========================================
# HEALTH
# =========================================

health:
	$(DC) ps

# =========================================
# BUILD (optional)
# =========================================

build-auth:
	cd auth && go build -o bin/auth cmd/main.go

build-user:
	cd user && go build -o bin/user cmd/main.go

build-image:
	cd image && go build -o bin/image cmd/main.go

build-processor:
	cd processor && go build -o bin/processor cmd/main.go

build: build-auth build-user build-image build-processor