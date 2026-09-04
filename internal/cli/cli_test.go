package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addozhang/jira-cli/internal/app"
)

func TestVersionCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader("")}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"version", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"schemaVersion":"1"`) {
		t.Fatalf("missing schema version: %s", out.String())
	}
}

func TestSearchRequiresJQLAndInstance(t *testing.T) {
	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader("")}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"search"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected missing jql error")
	}

	cmd = NewRootCommand(rt)
	cmd.SetArgs([]string{"search", "--jql", "project = PROJ"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected missing instance error")
	}
}

func TestIssueCommentRequiresBody(t *testing.T) {
	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader("")}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"issue", "comment", "PROJ-1", "--instance", "prod"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected missing body error")
	}
}

func TestFakeJiraIntegrationThroughCLI(t *testing.T) {
	var sawComment, sawAssign bool
	var assigned map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing bearer token for %s %s: %q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/2/myself":
			_ = json.NewEncoder(w).Encode(map[string]any{"name": "agent", "displayName": "Agent User"})
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/2/issue/PROJ-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"key": "PROJ-1", "fields": map[string]any{"summary": "Fake Jira issue", "status": map[string]any{"name": "Open"}, "issuetype": map[string]any{"name": "Bug"}, "project": map[string]any{"key": "PROJ"}, "reporter": map[string]any{"displayName": "Reporter", "name": "reporter1"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/2/search":
			if got := r.URL.Query().Get("jql"); got != "project = PROJ" {
				t.Fatalf("jql=%q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"startAt": 0, "maxResults": 50, "total": 1, "issues": []any{map[string]any{"key": "PROJ-1", "fields": map[string]any{"summary": "Fake Jira issue", "status": map[string]any{"name": "Open"}, "issuetype": map[string]any{"name": "Bug"}, "project": map[string]any{"key": "PROJ"}}}}})
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/2/issue/PROJ-1/comment":
			var payload map[string]string
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if payload["body"] != "hello from fake integration" {
				t.Fatalf("comment body=%q", payload["body"])
			}
			sawComment = true
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "10001", "body": payload["body"], "created": "2026-06-27T00:00:00.000+0000"})
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/2/issue/PROJ-1/comment":
			_ = json.NewEncoder(w).Encode(map[string]any{"startAt": 0, "maxResults": 50, "total": 1, "comments": []any{map[string]any{"id": "10001", "body": "first comment", "author": map[string]any{"displayName": "Agent User", "name": "agent"}, "created": "2026-06-27T00:00:00.000+0000", "updated": "2026-06-27T00:01:00.000+0000"}}})
		case r.Method == http.MethodPut && r.URL.Path == "/rest/api/2/issue/PROJ-1/assignee":
			assigned = map[string]string{}
			_ = json.NewDecoder(r.Body).Decode(&assigned)
			sawAssign = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")
	whoami := runCLI(t, credentialsPath, "", "auth", "whoami", "prod", "-o", "json")
	if !strings.Contains(whoami, `"displayName":"Agent User"`) {
		t.Fatalf("unexpected whoami output: %s", whoami)
	}
	issue := runCLI(t, credentialsPath, "", "issue", "get", server.URL+"/browse/PROJ-1", "-o", "json")
	if !strings.Contains(issue, `"key":"PROJ-1"`) || !strings.Contains(issue, `"summary":"Fake Jira issue"`) {
		t.Fatalf("unexpected issue output: %s", issue)
	}
	search := runCLI(t, credentialsPath, "", "search", "--jql", "project = PROJ", "--instance", "prod", "-o", "json")
	if !strings.Contains(search, `"total":1`) || !strings.Contains(search, `"key":"PROJ-1"`) {
		t.Fatalf("unexpected search output: %s", search)
	}
	comment := runCLI(t, credentialsPath, "", "issue", "comment", "PROJ-1", "--instance", "prod", "--body", "hello from fake integration", "-o", "json")
	if !sawComment || !strings.Contains(comment, `"id":"10001"`) {
		t.Fatalf("unexpected comment output: %s", comment)
	}
	commentsByURL := runCLI(t, credentialsPath, "", "issue", "comments", server.URL+"/browse/PROJ-1", "-o", "json")
	if !strings.Contains(commentsByURL, `"issueKey":"PROJ-1"`) || !strings.Contains(commentsByURL, `"author":"Agent User"`) {
		t.Fatalf("unexpected comments-by-url output: %s", commentsByURL)
	}
	commentsByKey := runCLI(t, credentialsPath, "", "issue", "comments", "PROJ-1", "--instance", "prod", "-o", "json")
	if !strings.Contains(commentsByKey, `"body":"first comment"`) {
		t.Fatalf("unexpected comments-by-key output: %s", commentsByKey)
	}

	issueWithReporter := runCLI(t, credentialsPath, "", "issue", "get", "PROJ-1", "--instance", "prod", "-o", "json")
	if !strings.Contains(issueWithReporter, `"reporter":"Reporter"`) || !strings.Contains(issueWithReporter, `"reporterUsername":"reporter1"`) {
		t.Fatalf("unexpected reporter username output: %s", issue)
	}
	if !strings.Contains(commentsByURL, `"authorUsername":"agent"`) {
		t.Fatalf("unexpected author username output: %s", commentsByURL)
	}

	assignByURL := runCLI(t, credentialsPath, "", "issue", "assign", server.URL+"/browse/PROJ-1", "reporter1", "-o", "json")
	if !sawAssign || assigned["name"] != "reporter1" || !strings.Contains(assignByURL, `"assignee":"reporter1"`) {
		t.Fatalf("unexpected assign-by-url output: %s assigned=%+v", assignByURL, assigned)
	}
	assigned = nil
	assignByKey := runCLI(t, credentialsPath, "", "issue", "assign", "PROJ-1", "agent", "--instance", "prod", "-o", "json")
	if assigned["name"] != "agent" || !strings.Contains(assignByKey, `"issueKey":"PROJ-1"`) {
		t.Fatalf("unexpected assign-by-key output: %s assigned=%+v", assignByKey, assigned)
	}
}

func TestFakeJiraAssignRejectedIsActionable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/rest/api/2/issue/PROJ-1/assignee" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"errorMessages": []string{"User 'nobody' cannot be assigned issues."}})
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")

	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader(""), CredentialsPath: credentialsPath}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"issue", "assign", "PROJ-1", "nobody", "--instance", "prod"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected assign rejection error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *app.Error, got %T", err)
	}
	if !strings.Contains(appErr.Message, "cannot be assigned") {
		t.Fatalf("expected jira message, got: %q", appErr.Message)
	}
	if !strings.Contains(appErr.Next, "jr issue get") {
		t.Fatalf("expected next step, got: %q", appErr.Next)
	}
}

func TestIssueGetDoesNotIncludeComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/PROJ-1" {
			t.Fatalf("unexpected request: %s", r.URL.String())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"key": "PROJ-1", "fields": map[string]any{"summary": "No comments here", "status": map[string]any{"name": "Open"}, "issuetype": map[string]any{"name": "Bug"}, "project": map[string]any{"key": "PROJ"}}})
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")
	issue := runCLI(t, credentialsPath, "", "issue", "get", "PROJ-1", "--instance", "prod", "-o", "json")
	if strings.Contains(issue, `"comments"`) {
		t.Fatalf("issue get unexpectedly included comments: %s", issue)
	}
}

func TestIssueAssignCommand(t *testing.T) {
	var assigned map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing bearer token for %s %s", r.Method, r.URL.Path)
		}
		if r.Method != http.MethodPut || r.URL.Path != "/rest/api/2/issue/PROJ-1/assignee" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		_ = json.NewDecoder(r.Body).Decode(&assigned)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")

	out := runCLI(t, credentialsPath, "", "issue", "assign", server.URL+"/browse/PROJ-1", "jdoe", "-o", "json")
	if assigned["name"] != "jdoe" {
		t.Fatalf("assignee not sent: %+v", assigned)
	}
	if !strings.Contains(out, `"assignee":"jdoe"`) || !strings.Contains(out, `"issueKey":"PROJ-1"`) || !strings.Contains(out, `"schemaVersion":"1"`) {
		t.Fatalf("unexpected assign confirmation: %s", out)
	}

	assigned = nil
	runCLI(t, credentialsPath, "", "issue", "assign", "PROJ-1", "jdoe", "--instance", "prod", "-o", "json")
	if assigned["name"] != "jdoe" {
		t.Fatalf("bare-key assign did not reach server: %+v", assigned)
	}
}

func TestIssueAssignBlankUsernameFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP request for blank username: %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")

	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader(""), CredentialsPath: credentialsPath}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"issue", "assign", "PROJ-1", "", "--instance", "prod", "-o", "json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected blank username error")
	}
}

func TestIssueAssignRawOutputNotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	credentialsPath := filepath.Join(t.TempDir(), "credentials")
	runCLI(t, credentialsPath, "token\n", "auth", "add", server.URL, "--alias", "prod", "-o", "json")

	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader(""), CredentialsPath: credentialsPath}
	cmd := NewRootCommand(rt)
	cmd.SetArgs([]string{"issue", "assign", server.URL + "/browse/PROJ-1", "jdoe", "-o", "raw"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected raw output error for assign")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || !strings.Contains(appErr.Message, "Raw output is not available") {
		t.Fatalf("unexpected raw error: %v", err)
	}
}

func runCLI(t *testing.T, credentialsPath, stdin string, args ...string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	rt := &Runtime{Out: &out, Err: &errOut, In: strings.NewReader(stdin), CredentialsPath: credentialsPath}
	cmd := NewRootCommand(rt)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("jr %s failed: %v\nstderr:\n%s\nstdout:\n%s", strings.Join(args, " "), err, errOut.String(), out.String())
	}
	return out.String()
}
