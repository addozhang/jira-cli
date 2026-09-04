# jr

URL-native Jira Server / Data Center CLI.

`jr` lets agents and terminal users read Jira issues and comments, search with
explicit JQL, add comments, and assign issues without opening the browser.

## Scope

- Server / Data Center only.
- Jira Cloud is not supported.
- First release: `auth`, `issue get`, `issue comments`, `issue comment`, `issue assign`, `search --jql`, `version`.
- `issue create/update/transition` and sprint/board commands are not in scope.

## Quick Start

```sh
jr auth add https://jira.example.com --alias prod
jr issue get https://jira.example.com/browse/PROJ-123 -o json
jr issue get PROJ-123 --instance prod
jr issue comments PROJ-123 --instance prod -o json
jr search --jql 'project = PROJ ORDER BY updated DESC' --instance prod
jr issue comment PROJ-123 --instance prod --body 'Investigating from terminal.'
jr issue assign PROJ-123 jdoe --instance prod
```

## Body Format

Comment bodies are Jira Server wiki markup and are passed through unchanged.
No Markdown conversion is performed.

## Assignee

`jr issue assign` accepts an exact username only — no display-name or email
matching. Take usernames from `issue get` (`reporterUsername`) or
`issue comments` (`authorUsername`) output.
