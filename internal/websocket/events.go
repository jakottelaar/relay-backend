package websocket

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/messages"
	"github.com/nats-io/nats.go"
)

type WebsocketEventHandler struct {
	nc      *nats.Conn
	manager *Manager
}

func NewWebsocketEventHandler(nc *nats.Conn, manager *Manager) *WebsocketEventHandler {
	return &WebsocketEventHandler{
		nc:      nc,
		manager: manager,
	}
}

func (h *WebsocketEventHandler) RegisterHandlers() error {
	_, err := h.nc.QueueSubscribe("messages.created", "websocket-workers", h.handleMessageCreated)
	if err != nil {
		log.Printf("Failed to subscribe to messages.created: %v", err)
		return err
	}
	return nil
}

func (h *WebsocketEventHandler) handleMessageCreated(msg *nats.Msg) {
	var message messages.CreateMessageResponse
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		log.Printf("Invalid messages.created event: %v", err)
		return
	}

	log.Printf("WebSocket received message.created event: %+v", message)

	payload := struct {
		Type    string                          `json:"type"`
		Message *messages.CreateMessageResponse `json:"message"`
		Sender  uuid.UUID                       `json:"sender"`
	}{
		Type: "MESSAGE_SENT",
		Message: &messages.CreateMessageResponse{
			ID:        message.ID,
			SenderID:  message.SenderID,
			ChannelID: message.ChannelID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		},
		Sender: message.SenderID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	// Broadcast to the channel using your Manager
	h.manager.BroadcastToChannel(message.ChannelID.String(), data)
}
