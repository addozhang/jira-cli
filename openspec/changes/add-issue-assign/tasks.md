## 1. Client

- [x] 1.1 Write failing test: `AssignIssue` sends `PUT /rest/api/2/issue/{key}/assignee` with `{"name": username}` and parses the 204 confirmation
- [x] 1.2 Write failing test: empty/whitespace username fails before any HTTP request
- [x] 1.3 Implement `JiraClient.AssignIssue(key, username)` returning `AssignResponse{SchemaVersion, IssueKey, Assignee, URL}`
- [x] 1.4 Write failing test: unassignable-user API error renders one-sentence message plus next step, with no follow-up API call
- [x] 1.5 Implement error mapping for rejected assignments per the error contract

## 2. Read-output username fields

- [x] 2.1 Write failing test: `issue get` maps `reporter.name` to `reporterUsername` (and omits when absent)
- [x] 2.2 Implement `ReporterUsername` in `IssueResponse` via `named.Name`; keep existing fields untouched
- [x] 2.3 Write failing test: comment entries map `author.name` to `authorUsername` (and omit when absent)
- [x] 2.4 Implement `AuthorUsername` in `CommentEntry` via `named.Name`
- [x] 2.5 Verify all existing tests still pass unchanged (additive-only contract)

## 3. Command wiring

- [x] 3.1 Write failing test: `jr issue assign <url> <username>` resolves the issue target and calls `AssignIssue`
- [x] 3.2 Write failing test: bare issue key requires `--instance` (same rule as `issue get`)
- [x] 3.3 Implement the `assign` subcommand with `--instance` flag in the issue command tree
- [x] 3.4 Verify `-o yaml|json|raw` rendering for the assign confirmation

## 4. Integration coverage

- [x] 4.1 Add fake Jira handler for the assignee endpoint (success, 400 unassignable, empty username)
- [x] 4.2 Add integration tests: URL target, bare-key target, error paths
- [x] 4.3 Run full suite with `-race`

## 5. Docs and skill

- [x] 5.1 Update `docs/schema.md`: `reporterUsername`, `authorUsername`, assign confirmation schema (additive, `schemaVersion` stays `"1"`)
- [x] 5.2 Update `README.md` command surface with `jr issue assign`
- [x] 5.3 Update `skills/jr-jira-cli/` with the assign workflow (comments/reporter → username → assign)
- [x] 5.4 Verify SPEC.md boundary note no longer lists assign as out of scope
