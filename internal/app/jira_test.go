package app

import (
	"encoding/json"
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
	comment, err := client.AddComment("PROJ-1", "hello")
	if err != nil || comment.ID != "100" {
		t.Fatalf("bad comment: %+v err=%v", comment, err)
	}
	search, err := client.Search("project = PROJ")
	if err != nil || search.Total != 1 {
		t.Fatalf("bad search: %+v err=%v", search, err)
	}
}
