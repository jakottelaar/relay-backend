package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarUrl string    `json:"avatar_url"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type SupabaseClient interface {
	GetUserByUsername(ctx context.Context, username string) (*Profile, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*Profile, error)
}

type supabaseClient struct {
	Url        string
	ApiKey     string
	httpClient *http.Client
}

func NewSupabaseClient(url, apiKey string) *supabaseClient {
	return &supabaseClient{
		Url:        url,
		ApiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *supabaseClient) GetUserByUsername(ctx context.Context, username string) (*Profile, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/rest/v1/profiles?username=eq.%s&select=id,username,avatar_url,created_at,updated_at", c.Url, username),
		nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("apikey", c.ApiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from Supabase: %w", err)
	}
	defer resp.Body.Close()

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to parse Supabase response: %w", err)
	}

	if len(profiles) == 0 {
		return nil, internal.NewNotFoundError("User not found")
	}

	return &profiles[0], nil
}

func (c *supabaseClient) GetUserByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/rest/v1/profiles?id=eq.%s&select=id,username,avatar_url,created_at,updated_at", c.Url, id.String()),
		nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("apikey", c.ApiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from Supabase: %w", err)
	}
	defer resp.Body.Close()

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to parse Supabase response: %w", err)
	}

	if len(profiles) == 0 {
		return nil, internal.NewNotFoundError("User not found")
	}

	return &profiles[0], nil
}

func (c *supabaseClient) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*Profile, error) {
	if len(userIDs) == 0 {
		return make(map[uuid.UUID]*Profile), nil
	}

	// Convert UUIDs to strings and join them with commas
	idStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		idStrings[i] = id.String()
	}
	idList := strings.Join(idStrings, ",")

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/rest/v1/profiles?id=in.(%s)&select=id,username,avatar_url,created_at,updated_at",
			c.Url, idList),
		nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("apikey", c.ApiKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users from Supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("got status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	profiles := []Profile{}
	err = json.Unmarshal(body, &profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert slice to map keyed by user ID for easy lookup
	profileMap := make(map[uuid.UUID]*Profile)
	for i := range profiles {
		// Parse the ID from string to UUID if needed
		id, err := uuid.Parse(profiles[i].ID.String())
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in profile: %w", err)
		}
		profileMap[id] = &profiles[i]
	}

	return profileMap, nil
}
