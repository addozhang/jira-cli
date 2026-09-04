## 1. Test First

- [x] 1.1 Add a failing Jira client test for `GET /rest/api/2/issue/{key}/comment` that verifies auth header, request path, response mapping, and raw output.
- [x] 1.2 Add a failing CLI fake Jira integration test for `jr issue comments` using both a full issue URL and a bare key with `--instance`.
- [x] 1.3 Add a failing CLI test proving `jr issue get` output does not include comments.

## 2. Jira Client

- [x] 2.1 Add comment page and comment entry schema types with `schemaVersion: "1"`.
- [x] 2.2 Implement `JiraClient.GetComments(issueKey string, raw bool)` using `GET /rest/api/2/issue/{key}/comment`.
- [x] 2.3 Map stable comment fields: `id`, `body`, `author`, `created`, and `updated`.

## 3. CLI Command

- [x] 3.1 Add `jr issue comments <url-or-key>` under the existing `issue` command.
- [x] 3.2 Reuse existing URL target resolution and `--instance URL|alias` handling.
- [x] 3.3 Support `-o yaml|json|raw` for comments output.

## 4. Documentation And Skill

- [x] 4.1 Update `README.md` quick start and scope to include `issue comments`.
- [x] 4.2 Update `docs/schema.md` with the comments page schema.
- [x] 4.3 Update `skills/jr-jira-cli/SKILL.md` to tell agents to use `issue comments` for discussion context.

## 5. Verification

- [x] 5.1 Run `go test ./...` and confirm the new failing tests pass after implementation.
- [x] 5.2 Run `make fmt`, `make test`, and `make lint` successfully.
- [x] 5.3 Run `openspec validate add-issue-comments` successfully.
