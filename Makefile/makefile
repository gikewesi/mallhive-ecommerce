DC=docker compose -f /home/miracle/project1/mallhive-ecommerce/docker-compose.yaml

.PHONY: build up down logs restart run

build:
	$(DC) build

up:
	$(DC) up -d

down:
	$(DC) down

logs:
	$(DC) logs -f

restart:
	make down && make up

run: build up logs
