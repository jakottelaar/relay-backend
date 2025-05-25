package infra

import (
	"fmt"

	"github.com/jakottelaar/relay-backend/config"
	"github.com/nats-io/nats.go"
)

func initializeNats(cfg *config.Config) (*nats.Conn, error) {
	nc, err := nats.Connect(cfg.NatsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}
