package messages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/channels"
)

type MessagesService interface {
	CreateMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error)
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
