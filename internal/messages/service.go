package messages

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/jakottelaar/relay-backend/internal/common"
	"github.com/nats-io/nats.go"
)

type MessagesService interface {
	CreateMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error)
	GetMessages(ctx context.Context, userID, channelID uuid.UUID, filters common.Filters) ([]*Message, common.Metadata, error)
	UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, content string) (*Message, error)
	DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error
}

type messagesService struct {
	messagesRepo    MessagesRepo
	channelsService channels.ChannelsService
	nc              *nats.Conn
}

func NewMessagesService(messagesRepo MessagesRepo, channelsService channels.ChannelsService, nc *nats.Conn) MessagesService {
	return &messagesService{
		messagesRepo:    messagesRepo,
		channelsService: channelsService,
		nc:              nc,
	}
}

func (s *messagesService) CreateMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error) {
	_, err := s.channelsService.GetDMChannelByID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	message, err := s.messagesRepo.SaveMessage(ctx, senderID, channelID, content)
	if err != nil {
		return nil, err
	}

	event := CreateMessageEvent{
		SenderID:  senderID,
		ChannelID: channelID,
		Content:   content,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, internal.NewInternalServerError("Failed to marshal create message event")
	}

	if err := s.nc.Publish(SubjectMessageCreate, eventData); err != nil {
		return nil, internal.NewInternalServerError("Failed to publish create message event")
	}

	return message, nil
}

func (s *messagesService) GetMessages(ctx context.Context, userID, channelID uuid.UUID, filters common.Filters) ([]*Message, common.Metadata, error) {
	_, err := s.channelsService.GetDMChannelByID(ctx, channelID)
	if err != nil {
		return nil, common.Metadata{}, err
	}

	messages, metadata, err := s.messagesRepo.FindMessages(ctx, userID, channelID, filters)
	if err != nil {
		return nil, metadata, err
	}

	return messages, metadata, nil
}

func (s *messagesService) UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, content string) (*Message, error) {
	message, err := s.messagesRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if message.SenderID != userID {
		return nil, internal.NewForbiddenError("You are not the sender of this message")
	}

	updatedMessage, err := s.messagesRepo.UpdateMessage(ctx, userID, messageID, content)
	if err != nil {
		return nil, err
	}

	event := UpdateMessageEvent{
		ID:        messageID,
		SenderID:  userID,
		ChannelID: message.ChannelID,
		Content:   content,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, internal.NewInternalServerError("Failed to marshal update message event")
	}

	if err := s.nc.Publish(SubjectMessageUpdate, eventData); err != nil {
		return nil, internal.NewInternalServerError("Failed to publish update message event")
	}

	return updatedMessage, nil
}

func (s *messagesService) DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error {
	message, err := s.messagesRepo.FindMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if message.SenderID != userID {
		return internal.NewForbiddenError("You are not the sender of this message")
	}

	err = s.messagesRepo.DeleteMessage(ctx, userID, messageID)
	if err != nil {
		return err
	}

	event := DeleteMessageEvent{
		ID:        messageID,
		ChannelID: message.ChannelID,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return internal.NewInternalServerError("Failed to marshal delete message event")
	}

	if err := s.nc.Publish(SubjectMessageDelete, eventData); err != nil {
		return internal.NewInternalServerError("Failed to publish delete message event")
	}

	return nil
}
