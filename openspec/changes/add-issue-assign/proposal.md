# Add Issue Assign

## Why

When triaging an issue, the person who should pick it up is usually already
named in the discussion — the reporter or a comment author. But `jr` cannot
close the loop: `issue get` and `issue comments` output only display names and
drop the username that Jira's assignee API requires, so the user must open the
browser and drag the ticket manually.

## What Changes

- Add a write command: `jr issue assign <url-or-key> <username>`.
- Call Jira Server / Data Center REST API v2:
  `PUT /rest/api/2/issue/{issueIdOrKey}/assignee` with `{"name": "<username>"}`.
- Expose the actionable identity in existing read output:
  - `issue get` adds `reporterUsername` alongside `reporter`.
  - `issue comments` adds `authorUsername` alongside `author`.
  - Additive fields only; `schemaVersion` stays `"1"`.
- Accept exact usernames only. No displayName/email fuzzy matching.
- On assign failure (user not assignable), surface Jira's message and suggest
  checking the username from `issue get` / `issue comments` output.
- Reuse the existing URL-as-identity resolver and `--instance URL|alias` rule
  for bare issue keys.
- Update docs, schema reference, agent skill, and fake Jira integration coverage.

This change deliberately stays minimal. It does not add a standalone
`jr user` command, unassign/default-assignee support, or any matching beyond
the exact username.

## Capabilities

### New Capabilities

- `issue-assign`: Assign a Jira issue to an exact username via CLI.

### Modified Capabilities

- `issue`: `issue get` structured output adds `reporterUsername` (additive).
- `issue-comments`: comment entries add `authorUsername` (additive).

## Impact

- **Commands**: adds `jr issue assign <url-or-key> <username> [--instance URL|alias]`.
- **Jira API**: uses `PUT /rest/api/2/issue/{issueIdOrKey}/assignee`.
- **Code**: Jira client (`AssignIssue`), issue command tree, schema types
  (`IssueEntry`, `CommentEntry`), named-user parsing (keep username, not just
  display name), docs, skill, and tests.
- **Compatibility**: no breaking changes; output changes are additive fields.
