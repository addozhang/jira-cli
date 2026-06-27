# jr e2e tests

`jr` does not require real Jira Server / Data Center e2e tests for the default
verification path.

Jira Server/Data Center requires a licensed instance, and this project does not
assume access to a trial license. Default CI and local verification use fake Jira
integration tests built with `httptest` instead. The primary CLI-level fake Jira
flow is `internal/cli.TestFakeJiraIntegrationThroughCLI`, which exercises:

- `jr auth add`
- `jr auth whoami`
- `jr issue get`
- `jr search --jql`
- `jr issue comment`

If a real Jira instance is available later, optional e2e tests can be added under
an explicit build tag and must be skipped unless all required environment
variables are set.
