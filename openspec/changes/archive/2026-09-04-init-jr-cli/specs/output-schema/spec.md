## ADDED Requirements

### Requirement: Structured output formats
The system SHALL support `-o yaml` and `-o json` for all commands that return structured data, with YAML as the default.

#### Scenario: Default output is YAML
- **WHEN** a user runs a structured command without `-o`
- **THEN** stdout is YAML and starts with `schemaVersion: "1"`

#### Scenario: JSON output is equivalent
- **WHEN** a user runs the same structured command with `-o json`
- **THEN** stdout contains the same fields as YAML encoded as JSON

### Requirement: Raw output format
The system SHALL support `-o raw` for commands where the underlying Jira response is useful to expose.

#### Scenario: Raw issue get returns Jira payload
- **WHEN** a user runs `jr issue get PROJ-123 -o raw --instance prod`
- **THEN** stdout contains the raw Jira response body for the issue request

### Requirement: Schema version envelope
Every structured response SHALL include `schemaVersion: "1"` at the top level.

#### Scenario: Issue get includes schema version
- **WHEN** `jr issue get` succeeds with structured output
- **THEN** the response includes top-level `schemaVersion: "1"`

### Requirement: Schema compatibility policy
The system SHALL document field stability tiers and versioning rules in `docs/schema.md`.

#### Scenario: Scripts can pin schema version
- **WHEN** a user reads `docs/schema.md`
- **THEN** the document explains that breaking changes require a new schema version and additive fields do not

### Requirement: Debug redaction
The system SHALL redact credentials from debug logs.

#### Scenario: Debug output hides authorization
- **WHEN** a user runs an online command with `--debug`
- **THEN** stderr does not include the stored PAT or an unredacted Authorization header
