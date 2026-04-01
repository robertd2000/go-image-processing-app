# =========================================
# DOCKER
# =========================================

DC=docker compose

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
# WAIT SERVICES
# =========================================

wait-db:
	@echo "Waiting for Postgres..."
	until $(DC) exec postgres-auth pg_isready >/dev/null 2>&1; do sleep 2; done
	until $(DC) exec postgres-user pg_isready >/dev/null 2>&1; do sleep 2; done
	until $(DC) exec postgres-image pg_isready >/dev/null 2>&1; do sleep 2; done
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
	$(DC) exec migrate sh -c 'migrate -path=/migrations/auth -database "$$AUTH_DB_URL" up'

auth-migrate-down:
	$(DC) exec migrate sh -c 'migrate -path=/migrations/auth -database "$$AUTH_DB_URL" down 1'

auth-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/auth -seq $(name)


user-migrate-up:
	$(DC) exec migrate sh -c 'migrate -path=/migrations/user -database "$$USER_DB_URL" up'

user-migrate-down:
	$(DC) exec migrate sh -c 'migrate -path=/migrations/user -database "$$USER_DB_URL" down 1'

user-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/user -seq $(name)


image-migrate-up:
	$(DC) exec migrate sh -c 'migrate -path=/migrations/image -database "$$IMAGE_DB_URL" up'

image-migrate-down:
	$(DC) exec migrate sh -c 'migrate -path=/migrations/image -database "$$IMAGE_DB_URL" down 1'

image-migrate-create:
	$(DC) exec migrate migrate create -ext sql -dir /migrations/image -seq $(name)

# =========================================
# KAFKA TOPICS
# =========================================

topics:
	$(DC) exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server kafka:9092 \
		--replication-factor 1 \
		--partitions 3 \
		--topic image.process

	$(DC) exec kafka kafka-topics \
		--create \
		--if-not-exists \
		--bootstrap-server kafka:9092 \
		--replication-factor 1 \
		--partitions 3 \
		--topic image.done

# =========================================
# BOOTSTRAP
# =========================================

bootstrap:
	make up
	make wait-db
	make wait-kafka
	make auth-migrate-up
	make user-migrate-up
	make image-migrate-up
	make topics

# =========================================
# RESET DATABASES
# =========================================

reset-auth-db:
	$(DC) exec postgres-auth psql -U auth_user -d auth_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset-user-db:
	$(DC) exec postgres-user psql -U user_user -d user_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset-image-db:
	$(DC) exec postgres-image psql -U image_user -d image_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

reset:
	make reset-auth-db
	make reset-user-db
	make reset-image-db
	make auth-migrate-up
	make user-migrate-up
	make image-migrate-up

# =========================================
# DEBUG
# =========================================

env-migrate:
	$(DC) exec migrate env

psql-auth:
	$(DC) exec postgres-auth psql -U auth_user -d auth_db

psql-user:
	$(DC) exec postgres-user psql -U user_user -d user_db

psql-image:
	$(DC) exec postgres-image psql -U image_user -d image_db

health:
	docker ps