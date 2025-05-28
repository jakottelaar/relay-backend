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
	_, err := h.nc.Subscribe(messages.SubjectMessageCreated, h.handleMessageCreated)
	if err != nil {
		return err
	}
	_, err = h.nc.Subscribe(messages.SubjectMessageUpdated, h.handleMessageUpdated)
	if err != nil {
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
		Type: "MESSAGE_CREATED",
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

func (h *WebsocketEventHandler) handleMessageUpdated(msg *nats.Msg) {
	var updatedMsg messages.UpdateMessageResponse
	if err := json.Unmarshal(msg.Data, &updatedMsg); err != nil {
		log.Printf("Invalid messages.updated event: %v", err)
		return
	}

	log.Printf("WebSocket received message.updated event: %+v", updatedMsg)

	payload := struct {
		Type    string                          `json:"type"`
		Message *messages.UpdateMessageResponse `json:"message"`
		Sender  uuid.UUID                       `json:"sender"`
	}{
		Type:    "MESSAGE_UPDATED",
		Message: &updatedMsg,
		Sender:  updatedMsg.SenderID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	log.Printf("Broadcasting message.updated event to channel %s", updatedMsg.ChannelID.String())

	h.manager.BroadcastToChannel(updatedMsg.ChannelID.String(), data)
}
