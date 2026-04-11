package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempAuth(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".es-cli.auth")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadAuth_SingleCluster(t *testing.T) {
	path := writeTempAuth(t, `{
		"prod": {
			"username": "admin",
			"password": "secret",
			"url": "http://10.0.0.1:9200"
		}
	}`)

	configs, err := LoadAuth(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(configs))
	}
	if configs[0].Name != "prod" {
		t.Fatalf("expected prod, got %s", configs[0].Name)
	}
	if configs[0].Username != "admin" {
		t.Fatalf("expected admin, got %s", configs[0].Username)
	}
	if configs[0].URL != "http://10.0.0.1:9200" {
		t.Fatalf("expected http://10.0.0.1:9200, got %s", configs[0].URL)
	}
}

func TestLoadAuth_MultiCluster(t *testing.T) {
	path := writeTempAuth(t, `{
		"staging": {"username": "u1", "password": "p1", "url": "http://staging:9200"},
		"prod":    {"username": "u2", "password": "p2", "url": "http://prod:9200"}
	}`)

	configs, err := LoadAuth(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(configs))
	}
	if configs[0].Name != "prod" {
		t.Fatalf("expected prod first (sorted), got %s", configs[0].Name)
	}
}

func TestLoadAuth_OldFormat(t *testing.T) {
	path := writeTempAuth(t, `{"elastic":"elastic"}`)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for old format")
	}
	if !strings.Contains(err.Error(), "old format") {
		t.Fatalf("expected old format error, got: %v", err)
	}
}

func TestLoadAuth_MissingUsername(t *testing.T) {
	path := writeTempAuth(t, `{"c1": {"username": "", "password": "p", "url": "http://x"}}`)
	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAuth_MissingURL(t *testing.T) {
	path := writeTempAuth(t, `{"c1": {"username": "u", "password": "p", "url": ""}}`)
	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAuth_EmptyFile(t *testing.T) {
	path := writeTempAuth(t, `{}`)
	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for empty")
	}
}

func TestLoadAuth_FileNotFound(t *testing.T) {
	_, err := LoadAuth("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAuth_InvalidJSON(t *testing.T) {
	path := writeTempAuth(t, `not json`)
	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultAuthPath(t *testing.T) {
	path := DefaultAuthPath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if !strings.HasSuffix(path, ".es-cli.auth") {
		t.Fatalf("expected .es-cli.auth suffix, got %s", path)
	}
}
