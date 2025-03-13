package channels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ChannelsService interface {
	GetDMChannel(ctx context.Context, userId, targetUserID uuid.UUID) (*Channel, error)
	CreateGroupChannel(ctx context.Context, userId uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error)
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

func (s *channelsService) CreateGroupChannel(ctx context.Context, ownerUserID uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error) {
	savedChannel, memberIDs, err := s.channelsRepo.SaveGroupChannel(ctx, ownerUserID, name, channelMemberIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("error saving group channel: %w", err)
	}

	return savedChannel, memberIDs, nil
}
