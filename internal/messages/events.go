package messages

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectMessageCreate  = "messages.create"
	SubjectMessageCreated = "messages.created"
	SubjectMessageUpdate  = "messages.update"
	SubjectMessageUpdated = "messages.updated"
)

type CreateMessageEvent struct {
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
}

type UpdateMessageEvent struct {
	ID       uuid.UUID `json:"id"`
	SenderID uuid.UUID `json:"sender_id"`
	Content  string    `json:"content"`
}

type MessagesEventHandler struct {
	nc  *nats.Conn
	svc MessagesService
}

func NewMessagesEventHandler(nc *nats.Conn, svc MessagesService) *MessagesEventHandler {
	return &MessagesEventHandler{
		nc:  nc,
		svc: svc,
	}
}

func (h *MessagesEventHandler) RegisterHandlers(ctx context.Context) error {
	// Change to subscribe instead of queue subscribe
	_, err := h.nc.Subscribe(SubjectMessageCreate, func(msg *nats.Msg) {
		if err := h.HandleCreateMessageEvent(ctx, msg); err != nil {
			log.Printf("Error handling create message event: %v", err)
			return
		}
	})

	if err != nil {
		return err
	}

	_, err = h.nc.Subscribe(SubjectMessageUpdate, func(msg *nats.Msg) {
		if err := h.HandleUpdateMessageEvent(ctx, msg); err != nil {
			log.Printf("Error handling update message event: %v", err)
			return
		}
	})

	if err != nil {
		return err
	}

	return nil
}

func (h *MessagesEventHandler) HandleCreateMessageEvent(ctx context.Context, msg *nats.Msg) error {
	var event CreateMessageEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return err
	}

	savedMessage, err := h.svc.CreateMessage(
		ctx,
		event.SenderID,
		event.ChannelID,
		event.Content,
	)

	if err != nil {
		return err
	}

	payload, err := json.Marshal(&CreateMessageResponse{
		ID:        savedMessage.ID,
		SenderID:  savedMessage.SenderID,
		ChannelID: savedMessage.ChannelID,
		Content:   savedMessage.Content,
		CreatedAt: savedMessage.CreatedAt,
	})
	if err != nil {
		log.Printf("Failed to marshal created message: %v", err)
		return err
	}

	if err := h.nc.Publish(SubjectMessageCreated, payload); err != nil {
		log.Printf("Failed to publish message created event: %v", err)
	}

	return nil
}

func (h *MessagesEventHandler) HandleUpdateMessageEvent(ctx context.Context, msg *nats.Msg) error {
	log.Printf("Received message update event: %s", msg.Subject)
	var event UpdateMessageEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return err
	}

	updatedMessage, err := h.svc.UpdateMessage(
		ctx,
		event.SenderID,
		event.ID,
		event.Content,
	)

	if err != nil {
		return err
	}

	payload, err := json.Marshal(&UpdateMessageResponse{
		ID:        updatedMessage.ID,
		SenderID:  updatedMessage.SenderID,
		ChannelID: updatedMessage.ChannelID,
		Content:   updatedMessage.Content,
		IsEdited:  updatedMessage.IsEdited,
		CreatedAt: updatedMessage.CreatedAt,
		UpdatedAt: updatedMessage.UpdatedAt,
	})
	if err != nil {
		log.Printf("Failed to marshal updated message: %v", err)
		return err
	}

	if err := h.nc.Publish(SubjectMessageUpdated, payload); err != nil {
		log.Printf("Failed to publish message updated event: %v", err)
	}

	return nil
}
