## Why

Developers who already drive Jenkins with `jk` and Confluence with `cfl` from the
terminal have no consistent, URL-native way to work with Jira Server / Data Center.
This change bootstraps `jr`: a Go CLI that makes the Jira URL you'd paste into a
browser the unit of identity, so reading, searching, and commenting on issues
feels identical to the rest of the toolchain — scriptable, agent-friendly, and
free of a heavyweight UI.

## What Changes

- Introduce the `jr` binary (`cmd/jr/`, module `github.com/addozhang/jira-cli`),
  built on Go 1.22+ and Cobra, mirroring the `jk`/`cfl` repository layout.
- **URL-as-identity** input model: every issue command accepts a Jira browse URL
  (`https://jira.example.com/browse/PROJ-123`), a REST URL, or a bare issue key
  (`PROJ-123`) disambiguated by `--instance URL|alias`.
  The host (plus optional reverse-proxy context path) selects the credential.
- **PAT Bearer auth**: `jr auth add/list/remove/whoami` stores Personal Access
  Tokens in `~/.config/jr/credentials` (TOML, mode `0600`), sent as
  `Authorization: Bearer <token>`. No username is stored or transmitted; tokens
  are never printed by any command.
- **Issue commands**: `jr issue get` and `jr issue comment`. Comments move as
  Jira Server **wiki markup**, passed through verbatim (no Markdown conversion —
  the caller owns the format).
- **Search**: `jr search --jql <query>` runs **JQL**, returning a single bounded
  page.
- **Stable self-owned schema**: `-o yaml|json|raw` (YAML default); every
  structured response begins with `schemaVersion: "1"`, with field stability tiers
  documented in `docs/schema.md`.
- **Operational ergonomics** shared with `jk`/`cfl`: global flags
  `-o/--output`, `--timeout` (30s), `--insecure`, `--debug`; `SSL_CERT_FILE`
  support; exit codes `0` (success) / `>=10` (any jr-level error); one-sentence
  actionable errors with a suggested next step; offline `jr version`.
- Ship a companion AI agent skill at `skills/jr-jira-cli/`, plus `SPEC.md`
  (engineering constitution) and GoReleaser + Homebrew tap distribution.

This is the foundational release. `issue create`, `issue update`, `issue
transition`, `issue assign`, and sprint/board (Jira Agile API) commands are
explicitly **out of scope** for this change. Jira **Cloud** is **out of scope**
(different auth model, REST v3, and ADF body format).

## Capabilities

### New Capabilities

- `cli-foundation`: The `jr` binary skeleton — root command, global flags
  (`-o/--output`, `--timeout`, `--insecure`, `--debug`), exit-code contract
  (`0` / `>=10`), actionable one-sentence errors, `SSL_CERT_FILE` handling, and
  the offline `version` command.
- `auth`: Credential lifecycle — `add/list/remove/whoami` for PAT Bearer tokens,
  storage at `~/.config/jr/credentials` (TOML, `0600`), host + context-path
  credential lookup, aliases, and the guarantee that tokens are never printed.
- `url-identity`: The shared URL/issue-key resolution model — accepted Jira URL
  shapes, bare-key + `--instance` selection rules, host/context-path
  normalization, and how a request resolves to a stored credential.
- `issue`: Issue operations — `get` and `comment`; comments as verbatim Jira
  Server wiki markup.
- `search`: JQL search via explicit `--jql`, returning a bounded page of issue
  summaries.
- `output-schema`: The self-owned output contract — `-o yaml|json|raw`,
  `schemaVersion: "1"` envelope, field stability tiers, and the versioning policy.

### Modified Capabilities

<!-- None. This is the initial change; no existing specs to modify. -->

## Impact

- **New repository scaffold**: `cmd/jr/`, `internal/` (config/auth, http client,
  url resolver, jira client, render, commands), `docs/schema.md`,
  `skills/jr-jira-cli/`, fake Jira integration tests, `SPEC.md`, `Makefile`,
  `.goreleaser.yaml`, `.golangci.yml`, `go.mod`/`go.sum`.
- **External dependencies**: Cobra (CLI), a TOML library (credentials), YAML
  encoder (output); transport on the standard-library `net/http`.
- **Jira Server / Data Center REST API** (`/rest/api/2/...`) is the integration
  surface; the Agile API is intentionally untouched in this change.
- **Filesystem**: creates and reads `~/.config/jr/credentials` (mode `0600`).
- **Distribution**: GoReleaser config + Homebrew tap `addozhang/tap`; MIT license.
