.SILENT:

.PHONY: build run stop down docker-build docker-up docker-down docker-logs


runapp:
	go run cmd/main.go


docker-build:
	docker compose -f docker/docker-compose.yaml build

docker-up:
	docker compose -f docker/docker-compose.yaml up -d

docker-down:
	docker compose -f docker/docker-compose.yaml down

docker-stop:
	docker compose -f docker/docker-compose.yaml stop

docker-start:
	docker compose -f docker/docker-compose.yaml start

docker-logs:
	docker compose -f docker/docker-compose.yaml logs -f

docker-restart:
	docker compose -f docker/docker-compose.yaml restart

docker-delete-images:
	docker rmi docker-subscriptions_service_app
	docker rmi postgres:17

delete-db:
	docker volume rm docker_postgres_data
