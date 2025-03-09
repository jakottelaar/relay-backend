package channels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ChannelsService interface {
	GetDMChannel(ctx context.Context, userId, targetUserID uuid.UUID) (*Channel, error)
}

type channelsService struct {
	channelsRepo ChannelsRepo
}

func NewChannelsService(channelsRepo ChannelsRepo) ChannelsService {
	return &channelsService{
		channelsRepo: channelsRepo,
	}
}

func (s *channelsService) GetDMChannel(ctx context.Context, userId, targetUserID uuid.UUID) (*Channel, error) {
	channel, err := s.channelsRepo.FindDMChannelByUserIDs(ctx, userId, targetUserID)
	if err != nil {
		// Any other error should be returned
		return nil, fmt.Errorf("error finding DM channel: %w", err)
	}

	if channel == nil {
		return s.channelsRepo.SaveDMChannel(ctx, userId, targetUserID)
	}

	// Channel was found, return it
	return channel, nil
}
