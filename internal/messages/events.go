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
)

type CreateMessageEvent struct {
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
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
	_, err := h.nc.QueueSubscribe(SubjectMessageCreate, "messages-workers", func(msg *nats.Msg) {
		if err := h.HandleCreateMessageEvent(ctx, msg); err != nil {
			log.Printf("Error handling create message event: %v", err)
			msg.Nak()
			return
		}
		msg.Ack()
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
