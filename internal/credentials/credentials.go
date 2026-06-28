package credentials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	refreshURL = "https://console.anthropic.com/v1/oauth/token"
	clientID   = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
)

type Credentials struct {
	AccessToken      string   `json:"accessToken"`
	RefreshToken     string   `json:"refreshToken"`
	ExpiresAt        int64    `json:"expiresAt"`
	Scopes           []string `json:"scopes,omitempty"`
	SubscriptionType string   `json:"subscriptionType,omitempty"`
	RateLimitTier    string   `json:"rateLimitTier,omitempty"`
}

type credentialsFile struct {
	ClaudeAiOauth Credentials `json:"claudeAiOauth"`
}

func Load(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var f credentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	if f.ClaudeAiOauth.AccessToken == "" {
		return nil, fmt.Errorf("accessToken is empty in credentials file")
	}

	return &f.ClaudeAiOauth, nil
}

func Refresh(path string) (*Credentials, error) {
	creds, err := Load(path)
	if err != nil {
		return nil, fmt.Errorf("load credentials for refresh: %w", err)
	}

	if creds.RefreshToken == "" {
		return nil, fmt.Errorf("refreshToken is empty, cannot refresh")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {creds.RefreshToken},
		"client_id":     {clientID},
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(refreshURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("call refresh endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh endpoint returned %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}

	creds.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		creds.RefreshToken = result.RefreshToken
	}
	if result.ExpiresIn > 0 {
		creds.ExpiresAt = time.Now().UnixMilli() + result.ExpiresIn*1000
	}

	f := credentialsFile{ClaudeAiOauth: *creds}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("write credentials: %w", err)
	}

	return creds, nil
}
