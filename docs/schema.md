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
`project`, `assignee`, `reporter`, `reporterUsername`, `description`, `labels`,
`updated`.

`reporter` is the display name; `reporterUsername` is the Jira username and is
the value accepted by `jr issue assign`.

## Comments

Stable page fields: `schemaVersion`, `issueKey`, `startAt`, `maxResults`,
`total`, `comments`.

Stable comment fields: `id`, `body`, `author`, `authorUsername`, `created`,
`updated`.

`author` is the display name; `authorUsername` is the Jira username and is the
value accepted by `jr issue assign`.

## Assign

Stable fields: `schemaVersion`, `issueKey`, `assignee`, `url`.

`assignee` echoes the username exactly as requested; Jira's `204 No Content`
response carries no body, so the value is not re-fetched.

## Search

Stable fields: `schemaVersion`, `jql`, `startAt`, `maxResults`, `total`, `issues`.

## Auth

Auth outputs never include tokens.
