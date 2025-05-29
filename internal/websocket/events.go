package websocket

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/messages"
	"github.com/jakottelaar/relay-backend/internal/relationships"
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
	_, err := h.nc.Subscribe(messages.SubjectMessageCreate, h.handleMessageCreated)
	if err != nil {
		return err
	}
	_, err = h.nc.Subscribe(messages.SubjectMessageUpdate, h.handleMessageUpdated)
	if err != nil {
		return err
	}

	_, err = h.nc.Subscribe(messages.SubjectMessageDelete, h.handleMessageDeleted)
	if err != nil {
		return err
	}

	_, err = h.nc.Subscribe(relationships.SubjectRelationshipCreate, h.handleRelationshipCreated)
	if err != nil {
		return err
	}
	return nil
}

func (h *WebsocketEventHandler) handleMessageCreated(msg *nats.Msg) {
	var message messages.CreateMessageEvent
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		log.Printf("Invalid messages.created event: %v", err)
		return
	}

	log.Printf("WebSocket received message.created event: %+v", message)

	payload := struct {
		Type    string                      `json:"type"`
		Message messages.CreateMessageEvent `json:"message"`
		Sender  uuid.UUID                   `json:"sender"`
	}{
		Type:    "MESSAGE_CREATE",
		Message: message,
		Sender:  message.SenderID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	h.manager.BroadcastToChannel(message.ChannelID.String(), data)
}

func (h *WebsocketEventHandler) handleMessageUpdated(msg *nats.Msg) {
	var updatedMessage messages.UpdateMessageEvent
	if err := json.Unmarshal(msg.Data, &updatedMessage); err != nil {
		log.Printf("Invalid messages.updated event: %v", err)
		return
	}

	log.Printf("WebSocket received message.updated event: %+v", updatedMessage)

	payload := struct {
		Type    string                      `json:"type"`
		Message messages.UpdateMessageEvent `json:"message"`
		Sender  uuid.UUID                   `json:"sender"`
	}{
		Type:    "MESSAGE_UPDATE",
		Message: updatedMessage,
		Sender:  updatedMessage.SenderID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	log.Printf("Broadcasting message.updated event to channel %s", updatedMessage.ChannelID.String())

	h.manager.BroadcastToChannel(updatedMessage.ChannelID.String(), data)
}

func (h *WebsocketEventHandler) handleMessageDeleted(msg *nats.Msg) {
	var deletedMessage messages.DeleteMessageEvent
	if err := json.Unmarshal(msg.Data, &deletedMessage); err != nil {
		log.Printf("Invalid messages.deleted event: %v", err)
		return
	}

	log.Printf("WebSocket received message.deleted event: %+v", deletedMessage)

	payload := struct {
		Type    string                      `json:"type"`
		Message messages.DeleteMessageEvent `json:"message"`
	}{
		Type:    "MESSAGE_DELETE",
		Message: deletedMessage,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	h.manager.BroadcastToChannel(deletedMessage.ChannelID.String(), data)
}

func (h *WebsocketEventHandler) handleRelationshipCreated(msg *nats.Msg) {
	var relationship relationships.CreateRelationshipEvent
	if err := json.Unmarshal(msg.Data, &relationship); err != nil {
		log.Printf("Invalid relationships.created event: %v", err)
		return
	}

	log.Printf("WebSocket received relationships.created event: %+v", relationship)

	payload := struct {
		Type string                                `json:"type"`
		Data relationships.CreateRelationshipEvent `json:"data"`
	}{
		Type: "RELATIONSHIP_FRIEND_REQUEST_CREATE",
		Data: relationship,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast payload: %v", err)
		return
	}

	h.manager.SendToUser(relationship.OtherUserID, data)
}
