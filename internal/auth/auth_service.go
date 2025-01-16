package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthService struct {
	BaseURL string
}

type ValidateTokenResponse struct {
	UserID int `json:"user_id"`
}

func NewAuthService(baseURL string) *AuthService {
	return &AuthService{BaseURL: baseURL}
}

func (a *AuthService) ValidateToken(ctx context.Context, token string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL+"/auth/validate-token", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("invalid token, status code: %d", resp.StatusCode)
	}

	var res ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return res.UserID, nil
}
