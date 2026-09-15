package app

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"
)

const keyringService = "jr"

var (
	keyringGet    = keyring.Get
	keyringSet    = keyring.Set
	keyringDelete = keyring.Delete
)

type CredentialsFile struct {
	Instances map[string]Credential `toml:"instances"`
	Aliases   map[string]string     `toml:"aliases"`
}

type Credential struct {
	Token  string `toml:"token"`
	Secure bool   `toml:"secure,omitempty"`
}

type CredentialStore struct {
	Path string
	Data CredentialsFile
}

func DefaultCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", WrapError("Could not find home directory", "Set HOME and try again.", err)
	}
	return filepath.Join(home, ".config", "jr", "credentials"), nil
}

func LoadCredentials(path string) (*CredentialStore, error) {
	store := &CredentialStore{Path: path, Data: CredentialsFile{Instances: map[string]Credential{}, Aliases: map[string]string{}}}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return store, nil
	} else if err != nil {
		return nil, WrapError("Could not read credentials file", "Check ~/.config/jr/credentials permissions.", err)
	}
	if _, err := toml.DecodeFile(path, &store.Data); err != nil {
		return nil, WrapError("Could not parse credentials file", "Fix ~/.config/jr/credentials TOML syntax.", err)
	}
	if store.Data.Instances == nil {
		store.Data.Instances = map[string]Credential{}
	}
	if store.Data.Aliases == nil {
		store.Data.Aliases = map[string]string{}
	}
	return store, nil
}

func (s *CredentialStore) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return WrapError("Could not create jr config directory", "Check permissions under ~/.config.", err)
	}
	tmp := s.Path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return WrapError("Could not write credentials file", "Check permissions under ~/.config/jr.", err)
	}
	enc := toml.NewEncoder(f)
	if err := enc.Encode(s.Data); err != nil {
		_ = f.Close()
		return WrapError("Could not encode credentials file", "Check credential values and try again.", err)
	}
	if err := f.Close(); err != nil {
		return WrapError("Could not close credentials file", "Check disk health and try again.", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return WrapError("Could not secure credentials file", "Set ~/.config/jr/credentials to mode 0600.", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return WrapError("Could not replace credentials file", "Check permissions under ~/.config/jr.", err)
	}
	return os.Chmod(s.Path, 0o600)
}

func (s *CredentialStore) Add(instanceURL, token, alias string, secure bool) (string, error) {
	key, err := NormalizeInstance(instanceURL)
	if err != nil {
		return "", err
	}
	if prev, ok := s.Data.Instances[key]; ok && prev.Secure {
		_ = keyringDelete(keyringService, key)
	}
	if secure {
		if err := keyringSet(keyringService, key, token); err != nil {
			return "", WrapError("Could not store the token in the OS keyring", "Run jr auth add "+key+" without --secure-storage to keep the token in the credentials file.", err)
		}
		s.Data.Instances[key] = Credential{Secure: true}
	} else {
		s.Data.Instances[key] = Credential{Token: token}
	}
	if alias != "" {
		s.Data.Aliases[alias] = key
	}
	return key, s.Save()
}

func (s *CredentialStore) Remove(value string) (bool, error) {
	removed := false
	secureKey := ""
	if key, ok := s.Data.Aliases[value]; ok {
		delete(s.Data.Aliases, value)
		if cred, ok := s.Data.Instances[key]; ok && cred.Secure {
			secureKey = key
		}
		delete(s.Data.Instances, key)
		removed = true
	} else if key, err := NormalizeInstance(value); err == nil {
		if cred, ok := s.Data.Instances[key]; ok {
			if cred.Secure {
				secureKey = key
			}
			delete(s.Data.Instances, key)
			removed = true
		}
		for alias, target := range s.Data.Aliases {
			if target == key {
				delete(s.Data.Aliases, alias)
			}
		}
	}
	if secureKey != "" {
		_ = keyringDelete(keyringService, secureKey)
	}
	return removed, s.Save()
}

func (s *CredentialStore) ResolveInstance(value string) (string, Credential, error) {
	if target, ok := s.Data.Aliases[value]; ok {
		cred, ok := s.Data.Instances[target]
		if !ok {
			return "", Credential{}, NewError("Alias points to a missing credential", "Run jr auth add for the alias again.")
		}
		return s.credential(target, cred)
	}
	key, err := NormalizeInstance(value)
	if err != nil {
		return "", Credential{}, err
	}
	cred, ok := s.Data.Instances[key]
	if !ok {
		return "", Credential{}, NewError("No credential configured for "+key, "Run jr auth add "+key+" first.")
	}
	return s.credential(key, cred)
}

func (s *CredentialStore) MatchURL(rawURL string) (string, Credential, error) {
	target, err := NormalizeInstance(rawURL)
	if err != nil {
		return "", Credential{}, err
	}
	keys := make([]string, 0, len(s.Data.Instances))
	for key := range s.Data.Instances {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, key := range keys {
		if instanceHasPrefix(target, key) {
			return s.credential(key, s.Data.Instances[key])
		}
	}
	return "", Credential{}, NewError("No credential configured for "+target, "Run jr auth add "+target+" first.")
}

func (s *CredentialStore) credential(key string, cred Credential) (string, Credential, error) {
	token, err := s.tokenFor(key, cred)
	if err != nil {
		return "", Credential{}, err
	}
	cred.Token = token
	return key, cred, nil
}

func (s *CredentialStore) tokenFor(key string, cred Credential) (string, error) {
	if !cred.Secure {
		return cred.Token, nil
	}
	token, err := keyringGet(keyringService, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", NewError("No token for "+key+" in the OS keyring", "Run jr auth add "+key+" --secure-storage again.")
	}
	if err != nil {
		return "", WrapError("Could not read the OS keyring", "Run jr auth add "+key+" without --secure-storage to keep the token in the credentials file.", err)
	}
	return token, nil
}

func (s *CredentialStore) SecureInstances() []string {
	secure := make([]string, 0)
	for key, cred := range s.Data.Instances {
		if cred.Secure {
			secure = append(secure, key)
		}
	}
	sort.Strings(secure)
	return secure
}

func NormalizeInstance(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", NewError("Invalid Jira instance URL: "+raw, "Use a full URL like https://jira.example.com.")
	}
	host := strings.ToLower(u.Host)
	if (u.Scheme == "https" && strings.HasSuffix(host, ":443")) || (u.Scheme == "http" && strings.HasSuffix(host, ":80")) {
		host = strings.TrimSuffix(strings.TrimSuffix(host, ":443"), ":80")
	}
	path := contextPath(u.Path)
	return u.Scheme + "://" + host + path, nil
}

func contextPath(path string) string {
	path = strings.TrimRight(path, "/")
	if path == "" {
		return ""
	}
	for _, marker := range []string{"/browse/", "/rest/api/2/"} {
		if idx := strings.Index(path+"/", marker); idx >= 0 {
			return strings.TrimRight(path[:idx], "/")
		}
	}
	return path
}

func instanceHasPrefix(target, key string) bool {
	if target == key {
		return true
	}
	return strings.HasPrefix(target, key+"/")
}

func RedactHeaderValue(value string) string {
	if value == "" {
		return ""
	}
	return "<redacted>"
}
