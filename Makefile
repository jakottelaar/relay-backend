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