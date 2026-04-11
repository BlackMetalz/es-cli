package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const DefaultAuthFile = ".es-cli.auth"

type ClusterConfig struct {
	Name     string
	Username string
	Password string
	URL      string
}

func DefaultAuthPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", DefaultAuthFile)
	}
	return filepath.Join(home, DefaultAuthFile)
}

func LoadAuth(path string) ([]ClusterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file %s: %w", path, err)
	}

	// Try parsing as multi-cluster format: {"name": {"username":"...", "password":"...", "url":"..."}}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid auth file format: %w", err)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("auth file is empty: no clusters configured")
	}

	// Detect old format: if any value is a plain string (not an object), it's the old format
	for name, val := range raw {
		trimmed := string(val)
		if len(trimmed) > 0 && trimmed[0] == '"' {
			return nil, fmt.Errorf(
				"auth file uses the old format {\"username\":\"password\"}.\n"+
					"Please update %s to the new multi-cluster format:\n\n"+
					"{\n"+
					"  \"%s\": {\n"+
					"    \"username\": \"your-username\",\n"+
					"    \"password\": \"your-password\",\n"+
					"    \"url\": \"http://localhost:9200\"\n"+
					"  }\n"+
					"}", path, name)
		}
	}

	var configs []ClusterConfig
	for name, val := range raw {
		var entry map[string]string
		if err := json.Unmarshal(val, &entry); err != nil {
			return nil, fmt.Errorf("invalid config for cluster %q: %w", name, err)
		}

		username := entry["username"]
		password := entry["password"]
		url := entry["url"]

		if username == "" {
			return nil, fmt.Errorf("cluster %q: username cannot be empty", name)
		}
		if password == "" {
			return nil, fmt.Errorf("cluster %q: password cannot be empty", name)
		}
		if url == "" {
			return nil, fmt.Errorf("cluster %q: url cannot be empty", name)
		}

		configs = append(configs, ClusterConfig{
			Name:     name,
			Username: username,
			Password: password,
			URL:      url,
		})
	}

	// Sort by name for stable ordering
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})

	return configs, nil
}
