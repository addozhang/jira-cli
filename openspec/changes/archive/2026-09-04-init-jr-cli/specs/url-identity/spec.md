## ADDED Requirements

### Requirement: Jira browse URL targets
The system SHALL accept Jira browse URLs as issue targets.

#### Scenario: Browse URL resolves issue key
- **WHEN** a user targets `https://jira.example.com/browse/PROJ-123`
- **THEN** the resolver extracts instance `https://jira.example.com` and issue key `PROJ-123`

### Requirement: Jira REST URL targets
The system SHALL accept Jira REST issue URLs as issue targets.

#### Scenario: REST URL resolves issue key
- **WHEN** a user targets `https://jira.example.com/rest/api/2/issue/PROJ-123`
- **THEN** the resolver extracts instance `https://jira.example.com` and issue key `PROJ-123`

### Requirement: Context-path instances
The system SHALL preserve path segments before the Jira route as the instance context path.

#### Scenario: Context path before browse route
- **WHEN** a user targets `https://jira.example.com/jira/browse/PROJ-123`
- **THEN** the resolver extracts instance `https://jira.example.com/jira` and issue key `PROJ-123`

### Requirement: Bare issue keys
The system SHALL accept bare issue keys only when an instance is provided with `--instance URL|alias`.

#### Scenario: Bare key with instance flag
- **WHEN** a user runs an issue command for `PROJ-123 --instance prod`
- **THEN** the resolver uses the `prod` alias as the instance and `PROJ-123` as the issue key

#### Scenario: Bare key without instance fails
- **WHEN** a user targets `PROJ-123` without `--instance`
- **THEN** the command fails with an actionable error asking for `--instance URL|alias`

### Requirement: URL normalization
The system SHALL strip trailing slashes and normalize default ports when resolving instances.

#### Scenario: Trailing slash is ignored
- **WHEN** a command targets `https://jira.example.com//browse/PROJ-123`
- **THEN** credential lookup uses the normalized instance key without a trailing slash
