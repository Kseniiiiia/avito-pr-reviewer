.PHONY: build up down test lint sqlc migrate

sqlc:
	docker run --rm -v $(PWD):/src -w /src kjconroy/sqlc generate

migrate:
	docker run --rm \
		--network host \
		-v $(PWD)/migrations:/migrations \
		migrate/migrate:latest \
		-path /migrations \
		-database "postgres://user:pass@localhost:5432/prdb?sslmode=disable" \
		up

build: sqlc
	docker build -t avito-pr-reviewer .

up: build migrate
	docker-compose up -d

down:
	docker-compose down -v

test:
	go test -v ./tests/...

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run