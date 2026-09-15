# jr

URL-native Jira Server / Data Center CLI.

`jr` lets agents and terminal users read Jira issues and comments, search with
explicit JQL, add comments, and assign issues without opening the browser.

## Disclaimer

> [!WARNING]
> This project started as a hands-on study of the Jira RESTful API and as a testbed for driving such a CLI from AI coding agents. Most commands are read-only, but a few perform writes (add comments, assign issues). Use with care in production environments and manage permissions tightly — give the stored credentials the least privilege they need.

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

## Credential storage

By default, tokens live in plain text at `~/.config/jr/credentials` (mode
`0600`). Pass `--secure-storage` to `jr auth add` to keep the token in the OS
keyring instead — macOS Keychain, Windows Credential Manager, or a freedesktop
Secret Service (GNOME Keyring / KWallet). With `--secure-storage` the
credentials file only holds a marker; the token never touches disk. Re-run
`jr auth add` without the flag to move a token back to the file.

## Body Format

Comment bodies are Jira Server wiki markup and are passed through unchanged.
No Markdown conversion is performed.

## Assignee

`jr issue assign` accepts an exact username only — no display-name or email
matching. Take usernames from `issue get` (`reporterUsername`) or
`issue comments` (`authorUsername`) output.
