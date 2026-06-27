# jr Engineering Constitution

`jr` is a URL-native Jira Server / Data Center CLI. It follows the same design
posture as `jk` and `cfl`: small Go binary, stable self-owned schemas, explicit
credentials, and behavior changes managed through OpenSpec.

## Boundaries

- Jira Server / Data Center only.
- Jira Cloud, REST v3, OAuth, and ADF are out of scope.
- First release supports auth, issue get, issue comment, explicit JQL search, and version.
- Issue create/update/transition/assign and Agile API commands are out of scope.

## Contracts

- Structured output includes `schemaVersion: "1"`.
- Tokens are never printed.
- Errors are one sentence plus a suggested next step.
- Exit code `0` means success; `>=10` means a `jr`-level failure.

## Development Mode

- Source-Driven Development: verify framework, library, and external API patterns against official documentation for the detected version before implementing them.
- Test-Driven Development: add or update tests for new behavior before writing the implementation, then keep tests green through refactors.
- Behavioral changes must have tests that assert observable outcomes rather than internal call sequences.
