package app

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func useMockKeyring(store map[string]string, fail bool) func() {
	origGet, origSet, origDelete := keyringGet, keyringSet, keyringDelete
	keyringGet = func(service, user string) (string, error) {
		if fail {
			return "", errors.New("keyring unavailable")
		}
		value, ok := store[user]
		if !ok {
			return "", keyring.ErrNotFound
		}
		return value, nil
	}
	keyringSet = func(service, user, password string) error {
		if fail {
			return errors.New("keyring unavailable")
		}
		store[user] = password
		return nil
	}
	keyringDelete = func(service, user string) error {
		delete(store, user)
		return nil
	}
	return func() { keyringGet, keyringSet, keyringDelete = origGet, origSet, origDelete }
}

func TestSecureStorageRoundTrip(t *testing.T) {
	mock := map[string]string{}
	restore := useMockKeyring(mock, false)
	defer restore()
	path := filepath.Join(t.TempDir(), "credentials")
	store, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.Add("https://jira.example.com", "secret", "prod", true)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "secret") {
		t.Fatalf("token leaked to credentials file: %s", raw)
	}
	loaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.SecureInstances(); len(got) != 1 || got[0] != key {
		t.Fatalf("secure instances = %v, want [%s]", got, key)
	}
	_, cred, err := loaded.ResolveInstance("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.Token != "secret" {
		t.Fatalf("token = %q, want %q", cred.Token, "secret")
	}
	removed, err := loaded.Remove("prod")
	if err != nil || !removed {
		t.Fatalf("remove failed: removed=%v err=%v", removed, err)
	}
	if _, ok := mock[key]; ok {
		t.Fatal("keyring entry was not deleted")
	}
}

func TestSecureStorageAddFailureKeepsFileClean(t *testing.T) {
	restore := useMockKeyring(map[string]string{}, true)
	defer restore()
	path := filepath.Join(t.TempDir(), "credentials")
	store, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Add("https://jira.example.com", "secret", "", true)
	if err == nil {
		t.Fatal("expected error when keyring is unavailable")
	}
	var addErr *Error
	if !errors.As(err, &addErr) || !strings.Contains(addErr.Next, "credentials file") {
		t.Fatalf("error lacks file fallback hint: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credentials file should not exist, got %v", err)
	}
}

func TestResolveMissingKeyringTokenIsActionable(t *testing.T) {
	restore := useMockKeyring(map[string]string{}, false)
	defer restore()
	store := &CredentialStore{Data: CredentialsFile{Instances: map[string]Credential{
		"https://jira.example.com": {Secure: true},
	}, Aliases: map[string]string{}}}
	_, _, err := store.ResolveInstance("https://jira.example.com")
	if err == nil {
		t.Fatal("expected error for missing keyring token")
	}
	var appErr *Error
	if !errors.As(err, &appErr) || !strings.Contains(appErr.Next, "secure-storage") {
		t.Fatalf("error lacks re-add hint: %v", err)
	}
}

func TestFileStorageReplacesSecureEntry(t *testing.T) {
	mock := map[string]string{}
	restore := useMockKeyring(mock, false)
	defer restore()
	path := filepath.Join(t.TempDir(), "credentials")
	store, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.Add("https://jira.example.com", "secret", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add("https://jira.example.com", "plain", "", false); err != nil {
		t.Fatal(err)
	}
	if len(mock) != 0 {
		t.Fatalf("keyring entry was not cleaned up: %v", mock)
	}
	loaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	_, cred, err := loaded.ResolveInstance(key)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Token != "plain" || cred.Secure {
		t.Fatalf("unexpected credential: %+v", cred)
	}
	if got := loaded.SecureInstances(); len(got) != 0 {
		t.Fatalf("secure instances = %v, want empty", got)
	}
}

func TestCredentialStoreSaveLoadRemoveAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	store, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.Add("https://jira.example.com:443/jira", "secret", "prod", false)
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
