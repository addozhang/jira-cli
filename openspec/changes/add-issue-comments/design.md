## Context

`jr` is currently agent-context-first, but `jr issue get` only returns summary
fields and `jr issue comment` only writes a new comment. Existing Jira discussion
threads are not available from the CLI, which forces agents back to the browser
or to manual copy/paste.

Source-Driven Development check: Jira Server/Data Center REST API v2 documents
`GET /rest/api/2/issue/{issueIdOrKey}/comment` as returning all comments for an
issue, with optional `expand`, `maxResults`, `orderBy`, and `startAt` query
parameters and a response containing `comments`, `maxResults`, `startAt`, and
`total`.

Official source:
https://developer.atlassian.com/server/jira/platform/rest/v11003/api-group-issue

## Goals / Non-Goals

**Goals:**

- Add a minimal read-only command for issue comments.
- Preserve URL-as-identity and bare-key `--instance` behavior.
- Return stable, agent-friendly structured output with `schemaVersion: "1"`.
- Keep fake Jira integration tests as the verification path.

**Non-Goals:**

- Comment update, delete, visibility editing, or moderation commands.
- Pagination, sorting, or expand flags in this change.
- Automatically including comments in `jr issue get`.
- Jira Cloud comment APIs or Atlassian Document Format support.

## Decisions

### Decision: Add `jr issue comments` as a separate command

Comments can be large and are not always needed when reading basic issue fields,
so `issue get` remains unchanged. `issue comments` makes the extra request only
when the user or agent explicitly asks for the discussion thread.

Alternatives considered:

- Add comments to `issue get` by default. Rejected because it changes existing
  output size and behavior.
- Add `--comments` to `issue get`. Rejected for now because a separate command is
  clearer for agents and keeps command behavior narrow.

### Decision: Use Jira's default comments page in the MVP

The command will not expose `startAt`, `maxResults`, `orderBy`, or `expand` flags
yet. It will call the documented comments endpoint without extra query parameters
and report Jira's returned pagination metadata.

Alternatives considered:

- Add pagination and ordering flags immediately. Rejected as a nice-to-have; the
  immediate user need is reading comments at all.

### Decision: Model only stable comment fields

The stable schema includes comment `id`, `body`, `author`, `created`, and
`updated`. `raw` remains the escape hatch for Jira-specific fields like
`renderedBody`, `visibility`, and `properties`.

Alternatives considered:

- Mirror the full Jira comment JSON. Rejected because the project owns a stable
  compact schema and raw output already covers the full server response.

## Risks / Trade-offs

- Large comment threads may be truncated by Jira defaults -> the response exposes
  `startAt`, `maxResults`, and `total`, so users can see whether Jira returned a
  partial page; pagination flags can be a later change.
- Comment author shapes vary across Jira versions -> map the best available
  display name from documented user fields and leave unknowns empty.
- Some users may expect comments in `issue get` -> docs and skill should point to
  `jr issue comments` for discussion context.

## Migration Plan

No migration is required. This is an additive command and does not change
existing schemas or command behavior.

## Open Questions

None.
