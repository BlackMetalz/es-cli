package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAuth_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`{"elastic":"elastic"}`), 0600)

	cred, err := LoadAuth(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.Username != "elastic" || cred.Password != "elastic" {
		t.Fatalf("expected elastic:elastic, got %s:%s", cred.Username, cred.Password)
	}
}

func TestLoadAuth_MissingFile(t *testing.T) {
	_, err := LoadAuth("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadAuth_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`not json`), 0600)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadAuth_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`{"user1":"pass1","user2":"pass2"}`), 0600)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for multiple entries")
	}
}

func TestLoadAuth_EmptyUsername(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`{"":"password"}`), 0600)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestLoadAuth_EmptyPassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`{"user":""}`), 0600)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestLoadAuth_EmptyObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth")
	os.WriteFile(path, []byte(`{}`), 0600)

	_, err := LoadAuth(path)
	if err == nil {
		t.Fatal("expected error for empty object")
	}
}

func TestDefaultAuthPath(t *testing.T) {
	path := DefaultAuthPath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}
