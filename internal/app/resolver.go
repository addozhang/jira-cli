package app

import (
	"net/url"
	"regexp"
	"strings"
)

var issueKeyRE = regexp.MustCompile(`^[A-Z][A-Z0-9]+-[0-9]+$`)

type IssueTarget struct {
	Instance string
	Key      string
}

func ResolveIssueTarget(target, instance string, store *CredentialStore) (IssueTarget, Credential, error) {
	if issueKeyRE.MatchString(target) {
		if instance == "" {
			return IssueTarget{}, Credential{}, NewError("Issue key requires an instance", "Pass --instance URL|alias.")
		}
		key, cred, err := store.ResolveInstance(instance)
		if err != nil {
			return IssueTarget{}, Credential{}, err
		}
		return IssueTarget{Instance: key, Key: target}, cred, nil
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return IssueTarget{}, Credential{}, NewError("Invalid issue target: "+target, "Use a Jira issue URL or PROJ-123 with --instance.")
	}
	issueKey := issueKeyFromPath(u.Path)
	if issueKey == "" {
		return IssueTarget{}, Credential{}, NewError("Could not find issue key in URL", "Use a Jira browse URL or REST issue URL.")
	}
	key, cred, err := store.MatchURL(target)
	if err != nil {
		return IssueTarget{}, Credential{}, err
	}
	return IssueTarget{Instance: key, Key: issueKey}, cred, nil
}

func issueKeyFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "browse" && i+1 < len(parts) && issueKeyRE.MatchString(parts[i+1]) {
			return parts[i+1]
		}
		if part == "issue" && i >= 2 && parts[i-2] == "api" && i+1 < len(parts) && issueKeyRE.MatchString(parts[i+1]) {
			return parts[i+1]
		}
	}
	return ""
}
