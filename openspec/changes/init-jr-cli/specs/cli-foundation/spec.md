## ADDED Requirements

### Requirement: Root CLI command
The system SHALL provide a `jr` binary with a Cobra-based command tree for Jira Server / Data Center workflows.

#### Scenario: Help lists core commands
- **WHEN** a user runs `jr --help`
- **THEN** the output lists `auth`, `issue`, `search`, and `version` commands

### Requirement: Global operational flags
The system SHALL support global flags `-o/--output`, `--timeout`, `--insecure`, and `--debug` across online commands.

#### Scenario: Output flag is accepted globally
- **WHEN** a user runs an online command with `-o json`
- **THEN** the command renders structured output as JSON using the same schema as YAML

#### Scenario: Timeout flag configures HTTP requests
- **WHEN** a user runs an online command with `--timeout 2m`
- **THEN** the command uses a two-minute per-request timeout

### Requirement: TLS configuration
The system SHALL honor `SSL_CERT_FILE` for custom CA bundles and SHALL support `--insecure` as an explicit opt-out from certificate verification.

#### Scenario: Custom CA bundle is configured
- **WHEN** `SSL_CERT_FILE` points to a readable PEM bundle
- **THEN** online commands use that bundle when verifying Jira TLS certificates

#### Scenario: Insecure mode is explicit
- **WHEN** a user passes `--insecure`
- **THEN** the command disables TLS verification and prints a warning to stderr

### Requirement: Exit code contract
The system SHALL exit `0` on success and `>=10` for any jr-level failure.

#### Scenario: Network error uses jr-level failure code
- **WHEN** Jira cannot be reached for an online command
- **THEN** the command exits with a code greater than or equal to `10`

### Requirement: Actionable errors
The system SHALL render failures as a one-sentence error plus a suggested next step.

#### Scenario: Authentication fails
- **WHEN** Jira rejects a request as unauthorized
- **THEN** stderr includes the failed action and suggests checking `jr auth whoami` or updating credentials

### Requirement: Version command
The system SHALL provide an offline `jr version` command that prints version, commit, and build date.

#### Scenario: Version runs without configuration
- **WHEN** no credentials file exists and a user runs `jr version`
- **THEN** the command succeeds without reading Jira configuration
