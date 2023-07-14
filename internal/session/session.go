package session

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type sessionInfo struct {
	UserID      string   `json:"user_id"`
	Memberships []string `json:"memberships"`
	Signed      bool     `json:"signed"`
}

// Parse will parse the given session information
func Parse(encoded string) (*sessionInfo, error) {

	if encoded == "" {
		return nil, fmt.Errorf("no session information provided")
	}

	jsonSession, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var session sessionInfo
	err = json.Unmarshal(jsonSession, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
