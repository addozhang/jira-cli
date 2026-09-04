package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultSearchLimit = 50

type JiraClient struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type IssueResponse struct {
	SchemaVersion    string   `json:"schemaVersion" yaml:"schemaVersion"`
	Key              string   `json:"key" yaml:"key"`
	URL              string   `json:"url" yaml:"url"`
	Summary          string   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Status           string   `json:"status,omitempty" yaml:"status,omitempty"`
	Type             string   `json:"type,omitempty" yaml:"type,omitempty"`
	Project          string   `json:"project,omitempty" yaml:"project,omitempty"`
	Assignee         string   `json:"assignee,omitempty" yaml:"assignee,omitempty"`
	Reporter         string   `json:"reporter,omitempty" yaml:"reporter,omitempty"`
	ReporterUsername string   `json:"reporterUsername,omitempty" yaml:"reporterUsername,omitempty"`
	Description      string   `json:"description,omitempty" yaml:"description,omitempty"`
	Labels           []string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Updated          string   `json:"updated,omitempty" yaml:"updated,omitempty"`
}

type CommentResponse struct {
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`
	IssueKey      string `json:"issueKey" yaml:"issueKey"`
	ID            string `json:"id,omitempty" yaml:"id,omitempty"`
	Body          string `json:"body,omitempty" yaml:"body,omitempty"`
	Created       string `json:"created,omitempty" yaml:"created,omitempty"`
}

type CommentsResponse struct {
	SchemaVersion string         `json:"schemaVersion" yaml:"schemaVersion"`
	IssueKey      string         `json:"issueKey" yaml:"issueKey"`
	StartAt       int            `json:"startAt" yaml:"startAt"`
	MaxResults    int            `json:"maxResults" yaml:"maxResults"`
	Total         int            `json:"total" yaml:"total"`
	Comments      []CommentEntry `json:"comments" yaml:"comments"`
}

type AssignResponse struct {
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`
	IssueKey      string `json:"issueKey" yaml:"issueKey"`
	Assignee      string `json:"assignee" yaml:"assignee"`
	URL           string `json:"url" yaml:"url"`
}

type CommentEntry struct {
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`
	Body           string `json:"body,omitempty" yaml:"body,omitempty"`
	Author         string `json:"author,omitempty" yaml:"author,omitempty"`
	AuthorUsername string `json:"authorUsername,omitempty" yaml:"authorUsername,omitempty"`
	Created        string `json:"created,omitempty" yaml:"created,omitempty"`
	Updated        string `json:"updated,omitempty" yaml:"updated,omitempty"`
}

type SearchResponse struct {
	SchemaVersion string         `json:"schemaVersion" yaml:"schemaVersion"`
	JQL           string         `json:"jql" yaml:"jql"`
	StartAt       int            `json:"startAt" yaml:"startAt"`
	MaxResults    int            `json:"maxResults" yaml:"maxResults"`
	Total         int            `json:"total" yaml:"total"`
	Issues        []IssueSummary `json:"issues" yaml:"issues"`
}

type IssueSummary struct {
	Key     string `json:"key" yaml:"key"`
	URL     string `json:"url" yaml:"url"`
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Status  string `json:"status,omitempty" yaml:"status,omitempty"`
	Type    string `json:"type,omitempty" yaml:"type,omitempty"`
}

func (c JiraClient) GetIssue(key string, raw bool) (any, error) {
	body, err := c.do(http.MethodGet, "/rest/api/2/issue/"+url.PathEscape(key), nil)
	if err != nil {
		return nil, err
	}
	if raw {
		return RawBody(body), nil
	}
	var wire jiraIssue
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, WrapError("Could not parse Jira issue response", "Try -o raw to inspect Jira's response.", err)
	}
	return mapIssue(c.BaseURL, wire), nil
}

func (c JiraClient) AddComment(key, body string) (CommentResponse, error) {
	payload, _ := json.Marshal(map[string]string{"body": body})
	respBody, err := c.do(http.MethodPost, "/rest/api/2/issue/"+url.PathEscape(key)+"/comment", payload)
	if err != nil {
		return CommentResponse{}, err
	}
	var wire struct {
		ID      string `json:"id"`
		Body    string `json:"body"`
		Created string `json:"created"`
	}
	if err := json.Unmarshal(respBody, &wire); err != nil {
		return CommentResponse{}, WrapError("Could not parse Jira comment response", "Run the command again with --debug for the raw exchange.", err)
	}
	return CommentResponse{SchemaVersion: SchemaVersion, IssueKey: key, ID: wire.ID, Body: wire.Body, Created: wire.Created}, nil
}

func (c JiraClient) GetComments(key string, raw bool) (any, error) {
	body, err := c.do(http.MethodGet, "/rest/api/2/issue/"+url.PathEscape(key)+"/comment", nil)
	if err != nil {
		return nil, err
	}
	if raw {
		return RawBody(body), nil
	}
	var wire struct {
		StartAt    int `json:"startAt"`
		MaxResults int `json:"maxResults"`
		Total      int `json:"total"`
		Comments   []struct {
			ID      string `json:"id"`
			Body    string `json:"body"`
			Created string `json:"created"`
			Updated string `json:"updated"`
			Author  *named `json:"author"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, WrapError("Could not parse Jira comments response", "Try -o raw to inspect Jira's response.", err)
	}
	comments := make([]CommentEntry, 0, len(wire.Comments))
	for _, comment := range wire.Comments {
		author := ""
		authorUsername := ""
		if comment.Author != nil {
			author = bestName(*comment.Author)
			authorUsername = comment.Author.Name
		}
		comments = append(comments, CommentEntry{ID: comment.ID, Body: comment.Body, Author: author, AuthorUsername: authorUsername, Created: comment.Created, Updated: comment.Updated})
	}
	return CommentsResponse{SchemaVersion: SchemaVersion, IssueKey: key, StartAt: wire.StartAt, MaxResults: wire.MaxResults, Total: wire.Total, Comments: comments}, nil
}

func (c JiraClient) Search(jql string) (SearchResponse, error) {
	query := url.Values{}
	query.Set("jql", jql)
	query.Set("maxResults", fmt.Sprint(defaultSearchLimit))
	body, err := c.do(http.MethodGet, "/rest/api/2/search?"+query.Encode(), nil)
	if err != nil {
		return SearchResponse{}, err
	}
	var wire struct {
		StartAt    int         `json:"startAt"`
		MaxResults int         `json:"maxResults"`
		Total      int         `json:"total"`
		Issues     []jiraIssue `json:"issues"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		return SearchResponse{}, WrapError("Could not parse Jira search response", "Run the command again with --debug for the raw exchange.", err)
	}
	issues := make([]IssueSummary, 0, len(wire.Issues))
	for _, issue := range wire.Issues {
		mapped := mapIssue(c.BaseURL, issue)
		issues = append(issues, IssueSummary{Key: mapped.Key, URL: mapped.URL, Summary: mapped.Summary, Status: mapped.Status, Type: mapped.Type})
	}
	return SearchResponse{SchemaVersion: SchemaVersion, JQL: jql, StartAt: wire.StartAt, MaxResults: wire.MaxResults, Total: wire.Total, Issues: issues}, nil
}

func (c JiraClient) AssignIssue(key, username string) (AssignResponse, error) {
	if strings.TrimSpace(username) == "" {
		return AssignResponse{}, NewError("Missing assignee username", "Pass an exact username, e.g. jr issue assign PROJ-123 jdoe.")
	}
	payload, _ := json.Marshal(map[string]string{"name": username})
	_, err := c.do(http.MethodPut, "/rest/api/2/issue/"+url.PathEscape(key)+"/assignee", payload)
	if err != nil {
		var appErr *Error
		if errors.As(err, &appErr) {
			detail := jiraErrorDetail(appErr.Next)
			if detail == "" {
				detail = appErr.Message
			}
			return AssignResponse{}, NewError(fmt.Sprintf("Could not assign %s to %s: %s", key, username, detail), "Check the exact username in jr issue get / jr issue comments output and try again.")
		}
		return AssignResponse{}, err
	}
	return AssignResponse{SchemaVersion: SchemaVersion, IssueKey: key, Assignee: username, URL: strings.TrimRight(c.BaseURL, "/") + "/browse/" + key}, nil
}

func jiraErrorDetail(next string) string {
	trimmed := strings.TrimSpace(next)
	if trimmed == "" || trimmed == httpFallbackNext {
		return ""
	}
	var wire struct {
		ErrorMessages []string `json:"errorMessages"`
	}
	if err := json.Unmarshal([]byte(trimmed), &wire); err == nil {
		if len(wire.ErrorMessages) > 0 {
			return wire.ErrorMessages[0]
		}
		return ""
	}
	return trimmed
}

func (c JiraClient) WhoAmI() (map[string]any, error) {
	body, err := c.do(http.MethodGet, "/rest/api/2/myself", nil)
	if err != nil {
		return nil, err
	}
	var wire map[string]any
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, WrapError("Could not parse Jira user response", "Run the command again with --debug for the raw exchange.", err)
	}
	wire["schemaVersion"] = SchemaVersion
	return wire, nil
}

func (c JiraClient) do(method, path string, body []byte) ([]byte, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	req, err := http.NewRequest(method, base+path, bytes.NewReader(body))
	if err != nil {
		return nil, WrapError("Could not create Jira request", "Check the Jira URL and try again.", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, WrapError("Could not reach Jira", "Check the instance URL, network, and TLS settings.", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapError("Could not read Jira response", "Try again or run with --debug.", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewError(fmt.Sprintf("Jira request failed with HTTP %d", resp.StatusCode), stringOrDefault(strings.TrimSpace(string(data)), httpFallbackNext))
	}
	return data, nil
}

const httpFallbackNext = "Check credentials, permissions, and the issue key."

type jiraIssue struct {
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary     string   `json:"summary"`
		Description string   `json:"description"`
		Labels      []string `json:"labels"`
		Updated     string   `json:"updated"`
		Status      named    `json:"status"`
		IssueType   named    `json:"issuetype"`
		Project     struct {
			Key string `json:"key"`
		} `json:"project"`
		Assignee *named `json:"assignee"`
		Reporter *named `json:"reporter"`
	} `json:"fields"`
}

type named struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func mapIssue(baseURL string, wire jiraIssue) IssueResponse {
	issueURL := strings.TrimRight(baseURL, "/") + "/browse/" + wire.Key
	assignee := ""
	if wire.Fields.Assignee != nil {
		assignee = bestName(*wire.Fields.Assignee)
	}
	reporter := ""
	reporterUsername := ""
	if wire.Fields.Reporter != nil {
		reporter = bestName(*wire.Fields.Reporter)
		reporterUsername = wire.Fields.Reporter.Name
	}
	return IssueResponse{SchemaVersion: SchemaVersion, Key: wire.Key, URL: issueURL, Summary: wire.Fields.Summary, Status: wire.Fields.Status.Name, Type: wire.Fields.IssueType.Name, Project: wire.Fields.Project.Key, Assignee: assignee, Reporter: reporter, ReporterUsername: reporterUsername, Description: wire.Fields.Description, Labels: wire.Fields.Labels, Updated: wire.Fields.Updated}
}

func bestName(value named) string {
	if value.DisplayName != "" {
		return value.DisplayName
	}
	return value.Name
}

func stringOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
