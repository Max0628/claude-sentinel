package credentials

import (
	"encoding/json"
	"fmt"
	"os"
)

type Credentials struct {
	AccessToken string `json:"accessToken"`
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
