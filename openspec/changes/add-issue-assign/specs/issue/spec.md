# Issue Capability Delta

## ADDED Requirements

### Requirement: Issue get exposes reporter username
The system SHALL include the reporter's Jira username as `reporterUsername` in
`jr issue get` structured output, in addition to the existing display-name
`reporter` field.

#### Scenario: Reporter username is mapped
- **WHEN** Jira returns an issue whose `reporter.name` is `jdoe`
- **THEN** structured output includes `reporterUsername: jdoe` alongside `reporter`

#### Scenario: Missing reporter omits the field
- **WHEN** Jira returns an issue without a reporter
- **THEN** `reporterUsername` is omitted from structured output

#### Scenario: Output remains additive
- **WHEN** a user runs `jr issue get <url-or-key>` in any output format
- **THEN** `schemaVersion` is still `"1"` and `reporterUsername` appears only as a new additional field
