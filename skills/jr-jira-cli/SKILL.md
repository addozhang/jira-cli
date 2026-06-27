# jr Jira CLI

Use `jr` when you need to inspect Jira Server / Data Center issues from the
terminal.

## Rules

- Prefer `-o json` for machine-readable inspection.
- Use the Jira URL as the target when available.
- For bare issue keys, pass `--instance URL|alias`.
- Search requires explicit `--jql` and `--instance`.
- Comments are Jira Server wiki markup; do not assume Markdown conversion.
- Never print or paste Personal Access Tokens.

## Examples

```sh
jr issue get https://jira.example.com/browse/PROJ-123 -o json
jr issue get PROJ-123 --instance prod -o json
jr search --jql 'project = PROJ ORDER BY updated DESC' --instance prod -o json
jr issue comment PROJ-123 --instance prod --body 'Checked from agent.'
```
