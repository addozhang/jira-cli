## Why

Agents can already read issue summary fields and add comments, but they cannot
inspect the existing discussion thread without opening Jira in a browser. This
leaves out critical context: prior decisions, reviewer feedback, and status
updates that often live only in comments.

## What Changes

- Add a read-only command: `jr issue comments <url-or-key>`.
- Reuse the existing URL-as-identity resolver and `--instance URL|alias` rule for
  bare issue keys.
- Call Jira Server / Data Center REST API v2:
  `GET /rest/api/2/issue/{issueIdOrKey}/comment`.
- Return a stable `schemaVersion: "1"` comments page with issue key, pagination
  metadata, and comment entries.
- Support `-o yaml|json|raw`; `raw` returns the Jira response body.
- Update docs, schema reference, agent skill, and fake Jira integration coverage.

This change deliberately stays read-only. It does not add comment update/delete,
pagination flags, sorting flags, rendered body modeling, or automatic comment
embedding in `jr issue get`.

## Capabilities

### New Capabilities

- `issue-comments`: Read Jira issue comments as a stable, agent-friendly schema.

### Modified Capabilities

<!-- None. This adds a new command without changing existing command behavior. -->

## Impact

- **Commands**: adds `jr issue comments <url-or-key> [--instance URL|alias]`.
- **Jira API**: uses `/rest/api/2/issue/{issueIdOrKey}/comment`.
- **Code**: updates Jira client, issue command tree, schema types, docs, skill,
  and tests.
- **Compatibility**: no breaking changes; existing `issue get`, `issue comment`,
  and `search` behavior is unchanged.
