package tests

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/auth"
	"github.com/jakottelaar/relay-backend/internal/supabase"
	"github.com/stretchr/testify/mock"
)

// MockAuthService implements the AuthService interface for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Authenticate(accessToken string) (*auth.AuthResponse, error) {
	args := m.Called(accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthResponse), args.Error(1)
}

func (m *MockAuthService) ExtractTokenFromHeader(header string) string {
	args := m.Called(header)
	return args.String(0)
}

// MockAuthMiddlewareProvider implements the AuthMiddlewareProvider interface
type MockAuthMiddlewareProvider struct {
	mock.Mock
}

func (m *MockAuthMiddlewareProvider) AuthMiddleware() gin.HandlerFunc {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(gin.HandlerFunc)
}

// MockSupabaseClient mocks the SupabaseClient interface
type MockSupabaseClient struct {
	mock.Mock
}

func (m *MockSupabaseClient) GetUserByUsername(ctx context.Context, username string) (*supabase.Profile, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil && args.Error(1) != nil {
		return nil, internal.NewNotFoundError("User not found")
	}

	if args.Get(0) != nil {
		return args.Get(0).(*supabase.Profile), nil
	}

	return nil, args.Error(1)
}

func (m *MockSupabaseClient) GetUserByID(ctx context.Context, id uuid.UUID) (*supabase.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil && args.Error(1) != nil {
		return nil, internal.NewNotFoundError("User not found")
	}

	if args.Get(0) != nil {
		return args.Get(0).(*supabase.Profile), nil
	}
	return args.Get(0).(*supabase.Profile), args.Error(1)
}

func (m *MockSupabaseClient) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*supabase.Profile, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*supabase.Profile), args.Error(1)
}
