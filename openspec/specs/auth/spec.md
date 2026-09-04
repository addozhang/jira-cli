### Requirement: Add PAT credential
The system SHALL allow users to store a Jira Server / Data Center Personal Access Token for an instance using `jr auth add <url> [--alias <name>]`.

#### Scenario: Token is entered via hidden prompt
- **WHEN** a user runs `jr auth add https://jira.example.com` and enters a token
- **THEN** the token is stored without being echoed to the terminal

#### Scenario: Alias is stored
- **WHEN** a user runs `jr auth add https://jira.example.com --alias prod`
- **THEN** later commands can refer to that instance as `prod`

### Requirement: Credential storage
The system SHALL store credentials in TOML at `~/.config/jr/credentials` with file mode `0600`.

#### Scenario: Credentials file is created securely
- **WHEN** the first credential is added
- **THEN** the credentials file exists with permissions readable and writable only by the owner

### Requirement: Credential lookup by normalized instance
The system SHALL key credentials by normalized `scheme://host[:port][/context]` and select the most specific context-path match for a Jira URL.

#### Scenario: Default HTTPS port is normalized
- **WHEN** a credential is stored for `https://jira.example.com`
- **THEN** a command targeting `https://jira.example.com:443/browse/PROJ-1` uses that credential

#### Scenario: Context path is more specific
- **WHEN** credentials exist for `https://jira.example.com` and `https://jira.example.com/team-a`
- **THEN** a command targeting `https://jira.example.com/team-a/browse/PROJ-1` uses the `team-a` credential

### Requirement: List credentials without secrets
The system SHALL list configured instances and aliases without printing tokens.

#### Scenario: Auth list redacts tokens
- **WHEN** a user runs `jr auth list`
- **THEN** stdout includes instance keys and aliases but no token values or token prefixes

### Requirement: Remove credentials idempotently
The system SHALL allow users to remove a stored credential by URL or alias using `jr auth remove <url-or-alias>`.

#### Scenario: Removing missing credential succeeds
- **WHEN** a user removes an instance that is not configured
- **THEN** the command exits successfully and reports that no credential was present

### Requirement: Verify current identity
The system SHALL provide `jr auth whoami <url-or-alias>` to verify a stored token against Jira.

#### Scenario: Token is valid
- **WHEN** Jira accepts the stored token
- **THEN** the command returns the authenticated user's stable schema without printing the token

### Requirement: Bearer token transport
The system SHALL send stored PATs as `Authorization: Bearer <token>` for online Jira requests.

#### Scenario: Request uses bearer auth
- **WHEN** an online command executes against a configured instance
- **THEN** the HTTP request includes a Bearer authorization header derived from the selected credential
