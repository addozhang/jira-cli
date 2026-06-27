package app

import "testing"

func TestNormalizeInstance(t *testing.T) {
	tests := map[string]string{
		"https://jira.example.com:443/browse/PROJ-1":        "https://jira.example.com",
		"http://jira.example.com:80/rest/api/2/issue/ABC-2": "http://jira.example.com",
		"https://jira.example.com/jira/browse/PROJ-1":       "https://jira.example.com/jira",
		"https://jira.example.com/jira/":                    "https://jira.example.com/jira",
	}
	for input, want := range tests {
		got, err := NormalizeInstance(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("NormalizeInstance(%q)=%q want %q", input, got, want)
		}
	}
}

func TestResolveIssueTarget(t *testing.T) {
	store := &CredentialStore{Data: CredentialsFile{Instances: map[string]Credential{"https://jira.example.com": {Token: "t"}}, Aliases: map[string]string{"prod": "https://jira.example.com"}}}
	target, _, err := ResolveIssueTarget("https://jira.example.com/browse/PROJ-1", "", store)
	if err != nil {
		t.Fatal(err)
	}
	if target.Key != "PROJ-1" || target.Instance != "https://jira.example.com" {
		t.Fatalf("unexpected target: %+v", target)
	}
	target, _, err = ResolveIssueTarget("PROJ-1", "prod", store)
	if err != nil {
		t.Fatal(err)
	}
	if target.Key != "PROJ-1" || target.Instance != "https://jira.example.com" {
		t.Fatalf("unexpected bare target: %+v", target)
	}
	if _, _, err := ResolveIssueTarget("PROJ-1", "", store); err == nil {
		t.Fatal("expected error for bare key without instance")
	}
	if _, _, err := ResolveIssueTarget("not-a-target", "", store); err == nil {
		t.Fatal("expected error for invalid target")
	}
}
