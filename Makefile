test:
	go test -v ./...

build:
	go build -o bin/ ./...

run:
	go run ./cmd/main.go

lint:
	golangci-lint run

clean:
	rm -rf bin/

migrate-up:
	migrate -path ./migrations -database "postgresql://postgres:postgres@localhost:6000/relay-db?sslmode=disable" up

migrate-down:
	migrate -path ./migrations -database "postgresql://postgres:postgres@localhost:6000/relay-db?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "Migration file created in ./migrations"
	@echo "Please fill in the up and down SQL statements in the migration file."
	@echo "Then run 'make migrate-up' to apply the migration."
	@echo "Or run 'make migrate-down' to revert the migration."

docker-build:
	docker build -t relay:latest .