COMPOSE = docker compose

DB_NAME = excel_template_mapper
DB_USER = app
DB_PASS = app

up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

db-up:
	$(COMPOSE) up -d mysql

db-down:
	$(COMPOSE) down

mysql:
	$(COMPOSE) exec mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME)

migrate-up:
	$(COMPOSE) exec -T mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < db/migrations/001_init.up.sql

migrate-down:
	$(COMPOSE) exec -T mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < db/migrations/001_init.down.sql
