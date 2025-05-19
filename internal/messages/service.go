package messages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/jakottelaar/relay-backend/internal/common"
)

type MessagesService interface {
	CreateMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error)
	GetMessages(ctx context.Context, userID, channelID uuid.UUID, filters common.Filters) ([]*Message, common.Metadata, error)
	UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, content string) (*Message, error)
}

type messagesService struct {
	messagesRepo    MessagesRepo
	channelsService channels.ChannelsService
}

func NewMessagesService(messagesRepo MessagesRepo, channelsService channels.ChannelsService) MessagesService {
	return &messagesService{
		messagesRepo:    messagesRepo,
		channelsService: channelsService,
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

	message.Content = content
	updatedMessage, err := s.messagesRepo.UpdateMessage(ctx, userID, messageID, content)
	if err != nil {
		return nil, err
	}

	return updatedMessage, nil
}
