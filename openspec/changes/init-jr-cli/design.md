## Context

`jr` is a new CLI in the same family as `jk` and `cfl`: small Go binaries that
turn the URL a user already has in their browser into a stable, scriptable CLI
interface. The current repository only contains OpenSpec scaffolding, so this
change establishes both the product contract and the implementation foundation.

The target platform is Jira Server / Data Center. Jira Cloud is intentionally
excluded because it has a different authentication model, REST API version, and
body representation. The first release is agent-context-first: authenticate,
resolve an issue URL or key, get issue context, search with JQL, and add a
comment when explicit command intent is provided.

The primary stakeholders are AI coding agents and terminal-first developers that
need reliable structured issue context plus an explicit way to leave comments.

## Goals / Non-Goals

**Goals:**

- Match the `jk`/`cfl` user experience: Go + Cobra, URL-as-identity, stable
  schemas, simple exit codes, actionable errors, and a companion agent skill.
- Support Jira Server / Data Center REST API v2 for core issue workflows.
- Preserve Jira Server wiki markup verbatim for issue descriptions and comments.
- Make credential lookup deterministic for bare hosts and reverse-proxy context
  paths, including aliases for compact commands.
- Keep the first implementation small enough to ship and test without Agile API
  or Jira administration scope.

**Non-Goals:**

- Jira Cloud support, OAuth flows, REST API v3, or Atlassian Document Format.
- Sprint, board, backlog, ranking, or other Jira Agile API commands.
- Issue create/update/transition/assign commands.
- Project, user, workflow, permission, field-configuration, or administration
  management beyond metadata needed to read issues and add comments.
- Markdown-to-wiki conversion or any rich-text transformation.

## Decisions

### Decision: Use Jira Server / Data Center REST API v2 only

`jr` will call `/rest/api/2/...` endpoints and document Jira Server / Data Center
as the supported platform.

Alternatives considered:

- Support Cloud and Server in one binary. Rejected for the initial release
  because Cloud uses different auth, v3 endpoints, and ADF bodies, which would
  dilute the simple `cfl`-style Server/DC posture.
- Abstract all Jira API variants behind an interface immediately. Rejected until
  there is a concrete Cloud requirement.

### Decision: Keep URL resolution in one shared package

All commands that accept an issue target will go through one resolver that parses
browse URLs, REST URLs, and bare issue keys with `--instance URL|alias`.
The resolver returns the normalized instance key, base REST URL, context path,
issue key, and selected credential.

Alternatives considered:

- Parse URLs separately inside each command. Rejected because context-path and
  alias rules are easy to drift and must match auth lookup exactly.

### Decision: Store PAT credentials as TOML with explicit instance keys

Credentials live in `~/.config/jr/credentials` with mode `0600`, keyed by
normalized `scheme://host[:port][/context]`. Optional aliases point at those
instance keys. Tokens are stored as Bearer PATs and are never printed.

Alternatives considered:

- OS keychain storage. Deferred because `jk`/`cfl` use plaintext config-file
  storage, which is portable, inspectable, and familiar to CLI users.
- Username + password or Basic auth. Rejected for the first release because PAT
  Bearer matches the `cfl` posture and avoids storing usernames.

### Decision: Treat Jira text bodies as opaque body text

Issue reads return Jira text fields as received, and `issue comment` accepts body
input from literals, files, or stdin and sends it verbatim to Jira Server fields
that expect wiki markup.

Alternatives considered:

- Convert Markdown to Jira wiki markup. Rejected because conversion is lossy and
  inconsistent, and LLM callers can generate the target format directly.
- Normalize or format wiki markup before sending. Rejected because the CLI should
  be a thin transport, not an editor.

### Decision: Own a stable CLI schema instead of mirroring raw Jira JSON

YAML and JSON outputs use a compact `schemaVersion: "1"` schema designed for
humans and scripts. `-o raw` returns the verbatim Jira response where that is
useful for debugging or unsupported fields.

Alternatives considered:

- Print Jira JSON by default. Rejected because Jira responses are large,
  inconsistent across versions, and poor for agent workflows.
- Hide raw output entirely. Rejected because raw output is useful when diagnosing
  Server/DC version differences.

### Decision: Limit writes to explicit comments

The initial write surface is only `jr issue comment <target> --body <input>`.
Create, update, transition, and assign are excluded from this change.

Alternatives considered:

- Ship full issue CRUD in the first release. Rejected because required custom
  fields and workflow variance would slow down the read-context MVP.
- Make the first release read-only. Rejected because agent workflows often need
  to leave a clear audit trail on the issue after inspecting it.

## Risks / Trade-offs

- Jira Server installations vary heavily in required fields and workflows -> this
  release avoids create/update/transition/assign and limits writes to comments.
- Jira Server/Data Center requires licensing for real instances -> project
  verification relies on fake Jira integration tests; real Jira e2e is not a
  requirement for this change.
- PAT Bearer support depends on Jira Server/DC version and configuration -> docs
  and errors will state Server/DC PAT Bearer as the supported auth path; other auth
  modes require a later OpenSpec change.
- Opaque wiki markup makes authoring less friendly for humans who prefer Markdown
  -> this is consistent with `cfl`; callers can generate or transform body content
  before invoking `jr`.
- `-o raw` can expose large Jira payloads -> tokens are still redacted from debug
  logs, and raw output only prints server responses, not local credentials.
- Search requires explicit JQL -> this avoids inventing a partial query builder,
  but users and agents must provide valid Jira Query Language.

## Migration Plan

This is a new project with no users and no persisted data to migrate. The rollout
is a normal first release: implement the scaffold, validate with unit/integration
tests, document schemas, add the agent skill, then publish via GoReleaser and the
Homebrew tap.

Rollback is removing the unreleased binary or reverting the initial commit. After
release, schema-breaking behavior requires a new schema version rather than a
silent change.

## Open Questions

None.
