run:
	docker compose -f docker-compose.local.yaml up

lint:
	golangci-lint run
