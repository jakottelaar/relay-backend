FROM golang:1.23.6-alpine AS build

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY go.mod go.sum ./

WORKDIR /app

# Download dependencies
RUN go mod download

# Copy application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy built binary
COPY --from=build /app/main .

CMD ["./main"]