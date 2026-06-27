package app

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type CredentialsFile struct {
	Instances map[string]Credential `toml:"instances"`
	Aliases   map[string]string     `toml:"aliases"`
}

type Credential struct {
	Token string `toml:"token"`
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

func (s *CredentialStore) Add(instanceURL, token, alias string) (string, error) {
	key, err := NormalizeInstance(instanceURL)
	if err != nil {
		return "", err
	}
	s.Data.Instances[key] = Credential{Token: token}
	if alias != "" {
		s.Data.Aliases[alias] = key
	}
	return key, s.Save()
}

func (s *CredentialStore) Remove(value string) (bool, error) {
	removed := false
	if key, ok := s.Data.Aliases[value]; ok {
		delete(s.Data.Aliases, value)
		delete(s.Data.Instances, key)
		removed = true
	} else if key, err := NormalizeInstance(value); err == nil {
		if _, ok := s.Data.Instances[key]; ok {
			delete(s.Data.Instances, key)
			removed = true
		}
		for alias, target := range s.Data.Aliases {
			if target == key {
				delete(s.Data.Aliases, alias)
			}
		}
	}
	return removed, s.Save()
}

func (s *CredentialStore) ResolveInstance(value string) (string, Credential, error) {
	if target, ok := s.Data.Aliases[value]; ok {
		cred, ok := s.Data.Instances[target]
		if !ok {
			return "", Credential{}, NewError("Alias points to a missing credential", "Run jr auth add for the alias again.")
		}
		return target, cred, nil
	}
	key, err := NormalizeInstance(value)
	if err != nil {
		return "", Credential{}, err
	}
	cred, ok := s.Data.Instances[key]
	if !ok {
		return "", Credential{}, NewError("No credential configured for "+key, "Run jr auth add "+key+" first.")
	}
	return key, cred, nil
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
			return key, s.Data.Instances[key], nil
		}
	}
	return "", Credential{}, NewError("No credential configured for "+target, "Run jr auth add "+target+" first.")
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
