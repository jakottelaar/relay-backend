package channels

import (
	"context"

	"github.com/google/uuid"
)

type ChannelsService interface {
	CreateChannel(ctx context.Context, name string, owner_id uuid.UUID, channel_type ChannelType) (*Channel, error)
}

type channelsService struct {
	channelsRepo ChannelsRepo
}

func NewChannelsService(channelsRepo ChannelsRepo) ChannelsService {
	return &channelsService{
		channelsRepo: channelsRepo,
	}
}

func (s *channelsService) CreateChannel(ctx context.Context, name string, owner_id uuid.UUID, channel_type ChannelType) (*Channel, error) {
	channel := &Channel{
		Name:        name,
		OwnerID:     owner_id,
		ChannelType: channel_type,
	}

	savedChannel, err := s.channelsRepo.SaveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	return savedChannel, nil
}
