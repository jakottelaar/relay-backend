package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarUrl string    `json:"avatar_url"`
	UpdatedAt string    `json:"updated_at"`
}

type SupabaseClient interface {
	GetUserByUsername(ctx context.Context, username string) (*Profile, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*Profile, error)
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
		fmt.Sprintf("%s/rest/v1/profiles?username=eq.%s&select=id,username,email,avatar_url,updated_at", c.Url, username),
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", c.ApiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &profiles[0], nil
}

func (c *supabaseClient) GetUserByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/rest/v1/profiles?id=eq.%s&select=id,username,email,avatar_url,updated_at", c.Url, id.String()),
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", c.ApiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &profiles[0], nil
}
