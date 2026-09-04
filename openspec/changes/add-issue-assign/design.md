# Design: Add Issue Assign

## Context

`jr` today is read-plus-comment: `issue get`, `issue comments`, `issue comment`,
`search`. Two facts shape this design:

1. The wire type `named` (internal/app/jira.go:226) already parses Jira's
   `name` (username) — but `bestName` keeps only `displayName`, so the identity
   the assignee API needs is parsed and then discarded.
2. `IssueResponse.Reporter`/`Assignee` and `CommentEntry.Author` store display
   names only, so callers have no stable path from "who said this" to
   "assign to them".

Confirmed intent (interview, 2026-09-04): the chain is
`comments/reporter → username → assign`. Exact usernames only; no standalone
user command; no unassign.

## Goals / Non-Goals

**Goals:**

- `jr issue assign <url-or-key> <username>` writes the assignee via REST v2.
- `issue get` and `issue comments` output carry the username needed as assign
  input (additive fields, `schemaVersion` stays `"1"`).
- Failure messages are actionable per the error contract.

**Non-Goals:**

- Standalone `jr user` command.
- Unassign / assign-to-default (`-1`) support.
- displayName/email fuzzy matching or any user search endpoint.
- Listing assignable users (`/user/assignable/search`).

## Decisions

### D1: Keep `bestName` for existing fields; add explicit username fields

`Reporter`/`Author` keep returning the display name. New additive fields
`ReporterUsername` (issue get) and `AuthorUsername` (comments) come straight
from `named.Name`.

- *Why not repurpose `Reporter`/`Author` to usernames?* Breaking change for
  existing consumers; schema contract says additive changes do not bump
  `schemaVersion`, and display names are the human-readable choice.
- *Why not nested objects (`reporter: {name, displayName}`)?* Larger shape
  change for the same information; flat additive fields match the existing
  flat-entry style.
- *Why no `assigneeUsername`?* Deliberate asymmetry, not an oversight: the
  confirmed workflow sources assign targets from reporter/comments only.
  Current-assignee username has no consumer in this chain; add it later with
  the same additive pattern if a need appears.

### D2: Assign via `PUT /rest/api/2/issue/{key}/assignee` with `{"name": ...}`

Server/DC REST v2 accepts `name` (username), `key`, or `accountId`. We send
`name` only — it is what our own read output exposes, and `accountId` is a
Cloud concept that is out of scope.

Jira returns `204 No Content` on success. The command renders a stable
confirmation: `schemaVersion`, `issueKey`, `assignee` (the requested username
echoed back — Jira's 204 has no body, so no re-fetch), and `url`.

Empty/whitespace username is rejected client-side before any HTTP call.

### D3: Error handling stays one-sentence + Next

A `400` from Jira (user not assignable / unknown user) surfaces Jira's message
as the one-liner; Next suggests checking the exact username in `jr issue get` /
`jr issue comments` output. No second API call to enumerate assignable users —
the confirmed workflow sources usernames from reporter/comments, not search.

### D4: Command placement — `jr issue assign`

Third subcommand under `issue`, same `--instance` flag pattern as `get`/
`comment`/`comments`, same `clientForIssue` resolution (URL or bare key).

## Risks / Trade-offs

- [Username contains characters needing escaping] → pass through
  `url.PathEscape` like every other path segment; body is JSON-encoded.
- [Display name and username both empty for a user] → fields are `omitempty`;
  caller falls back to `Next` guidance, no synthetic values.
- [Additive fields confuse old skill/docs consumers] → update
  `docs/schema.md` and `skills/jr-jira-cli` in the same change; fields are
  documented as stable additions.

## Migration Plan

None. Additive output fields and one new command; rollback is reverting the
commit.

## Open Questions

None. Scope was fixed by the confirmed intent interview.
