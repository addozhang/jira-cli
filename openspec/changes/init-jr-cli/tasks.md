## 1. Repository Foundation

- [x] 1.1 Create Go module `github.com/addozhang/jira-cli` and `cmd/jr` entrypoint.
- [x] 1.2 Add Cobra root command with global flags `-o/--output`, `--timeout`, `--insecure`, and `--debug`.
- [x] 1.3 Add `jr version` command wired to build-time version, commit, and date variables.
- [x] 1.4 Add `Makefile` targets for build, test, lint, fmt, and release snapshot.
- [x] 1.5 Add baseline `.gitignore`, `.golangci.yml`, `.goreleaser.yaml`, `LICENSE`, and `SPEC.md` mirroring the jk/cfl project conventions.

## 2. Shared Runtime

- [x] 2.1 Implement typed application errors with one-sentence messages, suggested next steps, and `>=10` exit handling.
- [x] 2.2 Implement HTTP client construction with timeout, `SSL_CERT_FILE`, `--insecure`, and debug logging.
- [x] 2.3 Redact `Authorization` and token-like values from all debug output.
- [x] 2.4 Implement output renderer for `yaml`, `json`, and `raw` formats.
- [x] 2.5 Add renderer tests proving YAML default, JSON equivalence, raw passthrough, and `schemaVersion: "1"` envelopes.

## 3. Authentication

- [x] 3.1 Implement credentials file load/save at `~/.config/jr/credentials` using TOML and mode `0600`.
- [x] 3.2 Implement normalized instance keys (`scheme://host[:port][/context]`) and alias mapping.
- [x] 3.3 Implement most-specific context-path credential lookup with segment-boundary matching.
- [x] 3.4 Add `jr auth add <url> [--alias <name>]` with hidden PAT prompt and token overwrite behavior.
- [x] 3.5 Add `jr auth list` without printing token values or token prefixes.
- [x] 3.6 Add idempotent `jr auth remove <url-or-alias>`.
- [x] 3.7 Add `jr auth whoami <url-or-alias>` using Jira `/rest/api/2/myself`.
- [x] 3.8 Add unit tests for credential storage permissions, aliases, normalization, and token redaction.

## 4. URL Identity Resolution

- [x] 4.1 Implement issue target parsing for Jira browse URLs (`/browse/PROJ-123`).
- [x] 4.2 Implement issue target parsing for Jira REST issue URLs (`/rest/api/2/issue/PROJ-123`).
- [x] 4.3 Implement context-path detection for instances mounted under prefixes such as `/jira`.
- [x] 4.4 Implement bare issue key resolution using required `--instance URL|alias`.
- [x] 4.5 Add actionable errors for bare keys used without `--instance`.
- [x] 4.6 Add resolver unit tests covering default ports, trailing slashes, context paths, aliases, and invalid targets.

## 5. Jira API Client

- [x] 5.1 Implement Jira REST v2 client primitives for authenticated GET, POST, PUT, and DELETE-like operations.
- [x] 5.2 Implement `/rest/api/2/issue/{key}` get with stable internal issue model mapping.
- [x] 5.3 Implement comment creation through `/rest/api/2/issue/{key}/comment` with verbatim body text.
- [x] 5.4 Implement explicit JQL search through `/rest/api/2/search` with a bounded `maxResults` value.
- [x] 5.5 Add client tests using `httptest` for auth headers, request paths, issue get mapping, comment payloads, search requests, server validation errors, and raw response handling.

## 6. Issue Commands

- [x] 6.1 Add `jr issue get <url-or-key>` with `--instance` support and `-o yaml|json|raw` output.
- [x] 6.2 Add body input helper supporting literal strings, `@path`, and `-` stdin for comment bodies.
- [x] 6.3 Add `jr issue comment <url-or-key> --body <input>`.
- [x] 6.4 Add command tests for required flags, no-request validation failures, resolver integration, and output formatting.

## 7. Search Command

- [x] 7.1 Add `jr search --jql <query> --instance <url-or-alias>`.
- [x] 7.2 Require explicit `--jql` and `--instance URL|alias` for search.
- [x] 7.3 Add tests for explicit JQL execution, instance selection, and search schema output.

## 8. Schema, Docs, And Agent Skill

- [x] 8.1 Write `docs/schema.md` covering schema versioning, field stability tiers, issue output, auth output, and search output.
- [x] 8.2 Add README quick start, install paths, command reference, Server/DC scope, Cloud non-goal, and scripting examples.
- [x] 8.3 Create `skills/jr-jira-cli/SKILL.md` for AI agents using `jr` safely from the terminal.
- [x] 8.4 Document wiki-markup body passthrough and the absence of Markdown conversion.
- [x] 8.5 Document PAT storage posture and token redaction guarantees.

## 9. Verification And Release Readiness

- [x] 9.1 Add integration tests with a fake Jira Server covering auth, issue get, comment, and search flows.
- [x] 9.2 Document that real Jira e2e is not required; fake Jira integration tests are the supported verification path.
- [x] 9.3 Run `make fmt`, `make test`, `make lint`, and `make release-snapshot` successfully.
- [x] 9.4 Validate OpenSpec artifacts and confirm all requirements are represented by implementation or tests.
