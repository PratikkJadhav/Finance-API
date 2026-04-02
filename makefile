run:
	go run cmd/server/main.go

docker-up:
	docker compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/finance_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/finance_db?sslmode=disable" down

build:
	go build -o bin/server cmd/server/main.go

seed:
	go run cmd/seed/main.go