package channels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/supabase"
)

type ChannelsService interface {
	GetDMChannel(ctx context.Context, currentUserID, targetUserID uuid.UUID) (*Channel, error)
	CreateGroupChannel(ctx context.Context, currentUserID uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error)
	GetAllChannels(ctx context.Context, currentUserID uuid.UUID) ([]*GetChannelResponse, error)
	GetDMChannelByID(ctx context.Context, channelID uuid.UUID) (*Channel, error)
}

type channelsService struct {
	channelsRepo   ChannelsRepo
	supabaseClient supabase.SupabaseClient
}

func NewChannelsService(channelsRepo ChannelsRepo, supabaseClient supabase.SupabaseClient) ChannelsService {
	return &channelsService{
		channelsRepo:   channelsRepo,
		supabaseClient: supabaseClient,
	}
}

func (s *channelsService) GetDMChannel(ctx context.Context, currentUserID, targetUserID uuid.UUID) (*Channel, error) {
	if currentUserID == targetUserID {
		return nil, internal.NewBadRequestError("Cannot create DM channel with self")
	}

	_, err := s.supabaseClient.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	channel, err := s.channelsRepo.FindDMChannelByUserIDs(ctx, currentUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("error finding DM channel: %w", err)
	}

	if channel == nil {
		return s.channelsRepo.SaveDMChannel(ctx, currentUserID, targetUserID)
	}

	return channel, nil
}

func (s *channelsService) CreateGroupChannel(ctx context.Context, ownerUserID uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error) {
	savedChannel, memberIDs, err := s.channelsRepo.SaveGroupChannel(ctx, ownerUserID, name, channelMemberIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("error saving group channel: %w", err)
	}

	return savedChannel, memberIDs, nil
}

func (s *channelsService) GetAllChannels(ctx context.Context, currentUserID uuid.UUID) ([]*GetChannelResponse, error) {
	channelWithMembers, err := s.channelsRepo.FindAllChannelsByUserID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("error finding all channels: %w", err)
	}

	// Gather all unique userIDs across all channels
	userIDSet := make(map[uuid.UUID]struct{})
	for _, cm := range channelWithMembers {
		for _, memberID := range cm.Members {
			userIDSet[memberID] = struct{}{}
		}
	}

	var allUserIDs []uuid.UUID
	for id := range userIDSet {
		allUserIDs = append(allUserIDs, id)
	}

	// Fetch all user profiles once
	userProfiles, err := s.supabaseClient.GetUsersByIDs(ctx, allUserIDs)
	if err != nil {
		return nil, fmt.Errorf("error fetching user profiles: %w", err)
	}

	// Construct the response
	var responses []*GetChannelResponse
	for _, cm := range channelWithMembers {
		var memberProfiles []supabase.Profile
		for _, memberID := range cm.Members {

			if memberID == currentUserID {
				continue
			}

			if profile, ok := userProfiles[memberID]; ok {
				memberProfiles = append(memberProfiles, supabase.Profile{
					ID:        profile.ID,
					Username:  profile.Username,
					Email:     profile.Email,
					AvatarUrl: profile.AvatarUrl,
					UpdatedAt: profile.UpdatedAt,
				})
			}
		}

		responses = append(responses, &GetChannelResponse{
			ID:             cm.Channel.ID,
			Name:           cm.Channel.Name,
			OwnerID:        cm.Channel.OwnerID,
			ChannelType:    cm.Channel.ChannelType,
			CreatedAt:      cm.Channel.CreatedAt,
			ChannelMembers: memberProfiles,
		})
	}

	return responses, nil
}

func (s *channelsService) GetDMChannelByID(ctx context.Context, channelID uuid.UUID) (*Channel, error) {
	channel, err := s.channelsRepo.FindDMChannelByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("error finding DM channel by ID: %w", err)
	}

	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	return channel, nil
}
