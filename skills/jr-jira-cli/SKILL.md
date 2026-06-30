---
name: jr-jira-cli
repo: addozhang/jira-cli
description: >
  URL-native Jira Server / Data Center CLI for AI agents. Use this skill when a
  user asks to inspect Jira issues, read issue comments, search with JQL, or add
  comments using the `jr` command.
---

# jr Jira CLI

Use `jr` when you need to inspect Jira Server / Data Center issues from the
terminal.

## Rules

- Prefer `-o json` for machine-readable inspection.
- Use the Jira URL as the target when available.
- For bare issue keys, pass `--instance URL|alias`.
- Use `jr issue comments <target> -o json` when you need the issue discussion
  thread or prior decisions from comments.
- Search requires explicit `--jql` and `--instance`.
- Comments are Jira Server wiki markup; do not assume Markdown conversion.
- Never print or paste Personal Access Tokens.

## Examples

```sh
jr issue get https://jira.example.com/browse/PROJ-123 -o json
jr issue get PROJ-123 --instance prod -o json
jr issue comments PROJ-123 --instance prod -o json
jr search --jql 'project = PROJ ORDER BY updated DESC' --instance prod -o json
jr issue comment PROJ-123 --instance prod --body 'Checked from agent.'
```
