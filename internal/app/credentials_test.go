package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialStoreSaveLoadRemoveAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	store, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.Add("https://jira.example.com:443/jira", "secret", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if key != "https://jira.example.com/jira" {
		t.Fatalf("unexpected key: %s", key)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("credentials mode=%o want 0600", got)
	}
	loaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	resolved, cred, err := loaded.ResolveInstance("prod")
	if err != nil {
		t.Fatal(err)
	}
	if resolved != key || cred.Token != "secret" {
		t.Fatalf("unexpected credential: %s %+v", resolved, cred)
	}
	removed, err := loaded.Remove("prod")
	if err != nil || !removed {
		t.Fatalf("remove failed: removed=%v err=%v", removed, err)
	}
}

func TestMostSpecificCredentialLookup(t *testing.T) {
	store := &CredentialStore{Data: CredentialsFile{Instances: map[string]Credential{
		"https://jira.example.com":      {Token: "root"},
		"https://jira.example.com/team": {Token: "team"},
	}, Aliases: map[string]string{}}}
	key, cred, err := store.MatchURL("https://jira.example.com/team/browse/PROJ-1")
	if err != nil {
		t.Fatal(err)
	}
	if key != "https://jira.example.com/team" || cred.Token != "team" {
		t.Fatalf("unexpected match: %s %+v", key, cred)
	}
}

func TestDebugTransportRedactsAuthorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	var debug bytes.Buffer
	client, err := NewHTTPClient(HTTPOptions{Debug: true, DebugOut: &debug})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")
	if _, err := client.Do(req); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(debug.String(), "secret") || !strings.Contains(debug.String(), "<redacted>") {
		t.Fatalf("authorization was not redacted: %s", debug.String())
	}
}
