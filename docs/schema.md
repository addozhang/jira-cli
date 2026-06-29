# jr Schema

Structured `yaml` and `json` output starts with:

```yaml
schemaVersion: "1"
```

Scripts should pin this value. Breaking changes require a new schema version;
additive fields do not.

## Stability

- `stable`: safe for scripts.
- `experimental`: may change without a schema bump.

## Issue

Stable fields: `schemaVersion`, `key`, `url`, `summary`, `status`, `type`,
`project`, `assignee`, `reporter`, `description`, `labels`, `updated`.

## Comments

Stable page fields: `schemaVersion`, `issueKey`, `startAt`, `maxResults`,
`total`, `comments`.

Stable comment fields: `id`, `body`, `author`, `created`, `updated`.

## Search

Stable fields: `schemaVersion`, `jql`, `startAt`, `maxResults`, `total`, `issues`.

## Auth

Auth outputs never include tokens.
