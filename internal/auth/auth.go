package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultAuthFile = ".es-cli.auth"

type Credential struct {
	Username string
	Password string
}

func DefaultAuthPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", DefaultAuthFile)
	}
	return filepath.Join(home, DefaultAuthFile)
}

func LoadAuth(path string) (Credential, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Credential{}, fmt.Errorf("failed to read auth file %s: %w", path, err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return Credential{}, fmt.Errorf("invalid auth file format: expected JSON object with {\"username\":\"password\"}: %w", err)
	}

	if len(raw) != 1 {
		return Credential{}, fmt.Errorf("invalid auth file format: expected exactly one entry {\"username\":\"password\"}, got %d entries", len(raw))
	}

	for username, password := range raw {
		if username == "" {
			return Credential{}, fmt.Errorf("invalid auth file: username cannot be empty")
		}
		if password == "" {
			return Credential{}, fmt.Errorf("invalid auth file: password cannot be empty")
		}
		return Credential{Username: username, Password: password}, nil
	}

	return Credential{}, fmt.Errorf("invalid auth file: no credentials found")
}
