DC=docker compose -f ../docker-compose.yaml

.PHONY: build up down logs restart run

build:
	$(DC) build

up:
	$(DC) up -d

down:
	$(DC) down

logs:
	$(DC) logs -f

push:
	$(DC) push

restart:
	make down && make up

run: build up logs

