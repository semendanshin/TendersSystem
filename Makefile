CONTAINER_NAME = tender_app

DATABASE_COMPOSE_FILE = database/db-compose.yaml
APP_COMPOSE_FILE = docker/app-compose.yaml

APP_WITH_DB_COMPOSE_FILE = docker/app-with-db-compose.yaml

APP_MIGRATE_UP_COMPOSE_FILE = docker/app-migrate-up-compose.yaml
APP_MIGRATE_UP_WITH_DB_COMPOSE_FILE = docker/app-migrate-up-with-db-compose.yaml

docker_with_db:
	docker compose -f $(DATABASE_COMPOSE_FILE) -f $(APP_COMPOSE_FILE) -f $(APP_WITH_DB_COMPOSE_FILE) up -d --build

docker_without_db:
	docker compose -f $(APP_COMPOSE_FILE) up -d --build

docker_down:
	docker compose -f $(DATABASE_COMPOSE_FILE) -f $(APP_COMPOSE_FILE) -f $(APP_WITH_DB_COMPOSE_FILE) down

migrate_up:
	goose -dir migrations postgres "$(DATABASE_URL)" up
