# jr

URL-native Jira Server / Data Center CLI.

`jr` lets agents and terminal users read Jira issues, search with explicit JQL,
and add comments without opening the browser.

## Scope

- Server / Data Center only.
- Jira Cloud is not supported.
- First release: `auth`, `issue get`, `issue comment`, `search --jql`, `version`.
- `issue create/update/transition/assign` and sprint/board commands are not in scope.

## Quick Start

```sh
jr auth add https://jira.example.com --alias prod
jr issue get https://jira.example.com/browse/PROJ-123 -o json
jr issue get PROJ-123 --instance prod
jr search --jql 'project = PROJ ORDER BY updated DESC' --instance prod
jr issue comment PROJ-123 --instance prod --body 'Investigating from terminal.'
```

## Body Format

Comment bodies are Jira Server wiki markup and are passed through unchanged.
No Markdown conversion is performed.
