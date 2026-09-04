package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJiraClientAuthIssueCommentSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header: %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.URL.Path == "/rest/api/2/issue/PROJ-1" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"key": "PROJ-1", "fields": map[string]any{"summary": "Bug", "status": map[string]any{"name": "Open"}, "issuetype": map[string]any{"name": "Bug"}, "project": map[string]any{"key": "PROJ"}}})
		case r.URL.Path == "/rest/api/2/issue/PROJ-1/comment" && r.Method == http.MethodPost:
			var payload map[string]string
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if payload["body"] != "hello" {
				t.Fatalf("bad comment body: %+v", payload)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "100", "body": payload["body"]})
		case r.URL.Path == "/rest/api/2/search" && r.Method == http.MethodGet:
			if !strings.Contains(r.URL.RawQuery, "jql=") {
				t.Fatalf("missing jql query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"startAt": 0, "maxResults": 50, "total": 1, "issues": []any{map[string]any{"key": "PROJ-1", "fields": map[string]any{"summary": "Bug"}}}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()
	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	issue, err := client.GetIssue("PROJ-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if issue.(IssueResponse).Key != "PROJ-1" {
		t.Fatalf("bad issue: %+v", issue)
	}
	if issue.(IssueResponse).ReporterUsername != "" {
		t.Fatalf("expected empty reporterUsername without reporter: %+v", issue)
	}
	comment, err := client.AddComment("PROJ-1", "hello")
	if err != nil || comment.ID != "100" {
		t.Fatalf("bad comment: %+v err=%v", comment, err)
	}
	search, err := client.Search("project = PROJ")
	if err != nil || search.Total != 1 {
		t.Fatalf("bad search: %+v err=%v", search, err)
	}
}

func TestJiraClientAssignIssue(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header: %q", r.Header.Get("Authorization"))
		}
		if r.Method != http.MethodPut || r.URL.Path != "/rest/api/2/issue/PROJ-1/assignee" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	confirmation, err := client.AssignIssue("PROJ-1", "jdoe")
	if err != nil {
		t.Fatal(err)
	}
	if payload["name"] != "jdoe" {
		t.Fatalf("bad assignee payload: %+v", payload)
	}
	if confirmation.SchemaVersion != SchemaVersion || confirmation.IssueKey != "PROJ-1" || confirmation.Assignee != "jdoe" {
		t.Fatalf("bad confirmation: %+v", confirmation)
	}
	if !strings.HasSuffix(confirmation.URL, "/browse/PROJ-1") {
		t.Fatalf("bad confirmation url: %s", confirmation.URL)
	}
}

func TestAssignIssueRejectsBlankUsername(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP request for blank username: %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	for _, username := range []string{"", "   "} {
		if _, err := client.AssignIssue("PROJ-1", username); err == nil {
			t.Fatalf("expected error for username %q", username)
		}
	}
}

func TestAssignIssueUnassignableErrorIsActionable(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"errorMessages": []string{"User 'jdoe' cannot be assigned issues."}})
	}))
	defer server.Close()
	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	_, err := client.AssignIssue("PROJ-1", "jdoe")
	if err == nil {
		t.Fatal("expected error for rejected assignment")
	}
	var appErr *Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *app.Error, got %T", err)
	}
	if !strings.Contains(appErr.Message, "cannot be assigned") {
		t.Fatalf("error should surface Jira message: %q", appErr.Message)
	}
	if !strings.Contains(appErr.Next, "jr issue get") || !strings.Contains(appErr.Next, "jr issue comments") {
		t.Fatalf("next step should point at username sources: %q", appErr.Next)
	}
	if requests != 1 {
		t.Fatalf("expected exactly 1 request, got %d", requests)
	}
}

func TestIssueAndCommentsExposeUsernames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/api/2/issue/PROJ-1" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"key": "PROJ-1", "fields": map[string]any{
				"summary":   "Bug",
				"status":    map[string]any{"name": "Open"},
				"issuetype": map[string]any{"name": "Bug"},
				"project":   map[string]any{"key": "PROJ"},
				"reporter":  map[string]any{"displayName": "Jane Doe", "name": "jdoe"},
			}})
		case r.URL.Path == "/rest/api/2/issue/PROJ-1/comment" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"startAt": 0, "maxResults": 50, "total": 2,
				"comments": []any{
					map[string]any{"id": "1", "body": "hi", "author": map[string]any{"displayName": "Jane Doe", "name": "jdoe"}, "created": "c", "updated": "u"},
					map[string]any{"id": "2", "body": "anonymous", "created": "c", "updated": "u"},
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()
	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	issue, err := client.GetIssue("PROJ-1", false)
	if err != nil {
		t.Fatal(err)
	}
	mapped := issue.(IssueResponse)
	if mapped.Reporter != "Jane Doe" || mapped.ReporterUsername != "jdoe" {
		t.Fatalf("bad reporter mapping: %+v", mapped)
	}
	page, err := client.GetComments("PROJ-1", false)
	if err != nil {
		t.Fatal(err)
	}
	comments := page.(CommentsResponse)
	if len(comments.Comments) != 2 {
		t.Fatalf("bad comment count: %+v", comments)
	}
	if comments.Comments[0].Author != "Jane Doe" || comments.Comments[0].AuthorUsername != "jdoe" {
		t.Fatalf("bad author mapping: %+v", comments.Comments[0])
	}
	if comments.Comments[1].Author != "" || comments.Comments[1].AuthorUsername != "" {
		t.Fatalf("expected empty author fields without author: %+v", comments.Comments[1])
	}
}

func TestJiraClientGetComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header: %q", r.Header.Get("Authorization"))
		}
		if r.Method != http.MethodGet || r.URL.Path != "/rest/api/2/issue/PROJ-1/comment" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"startAt":    0,
			"maxResults": 50,
			"total":      1,
			"comments": []any{map[string]any{
				"id":      "10001",
				"body":    "first comment",
				"created": "2026-06-27T00:00:00.000+0000",
				"updated": "2026-06-27T00:01:00.000+0000",
				"author":  map[string]any{"displayName": "Agent User", "name": "agent"},
			}},
		})
	}))
	defer server.Close()

	client := JiraClient{BaseURL: server.URL, Token: "token", HTTP: server.Client()}
	comments, err := client.GetComments("PROJ-1", false)
	if err != nil {
		t.Fatal(err)
	}
	page := comments.(CommentsResponse)
	if page.SchemaVersion != SchemaVersion || page.IssueKey != "PROJ-1" || page.Total != 1 {
		t.Fatalf("bad comments page: %+v", page)
	}
	if len(page.Comments) != 1 || page.Comments[0].Author != "Agent User" || page.Comments[0].Body != "first comment" {
		t.Fatalf("bad comment mapping: %+v", page.Comments)
	}
	raw, err := client.GetComments("PROJ-1", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw.(RawBody)), `"comments"`) {
		t.Fatalf("raw comments did not include Jira body: %s", raw)
	}
}
