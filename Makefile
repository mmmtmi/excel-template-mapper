COMPOSE = docker compose

DB_NAME = excel_template_mapper
DB_USER = app
DB_PASS = app

db-up:
	$(COMPOSE) up -d

db-down:
	$(COMPOSE) down

mysql:
	$(COMPOSE) exec mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME)

migrate-up:
	$(COMPOSE) exec -T mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < db/migrations/001_init.up.sql

migrate-down:
	$(COMPOSE) exec -T mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < db/migrations/001_init.down.sql