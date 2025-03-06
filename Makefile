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