# Issue Comments Capability Delta

## ADDED Requirements

### Requirement: Comment entries expose author username
The system SHALL include the comment author's Jira username as
`authorUsername` in each `jr issue comments` entry, in addition to the
existing display-name `author` field.

#### Scenario: Author username is mapped
- **WHEN** Jira returns a comment whose `author.name` is `jdoe`
- **THEN** that comment entry includes `authorUsername: jdoe` alongside `author`

#### Scenario: Missing author omits the field
- **WHEN** Jira returns a comment without an author
- **THEN** `authorUsername` is omitted from that entry

#### Scenario: Output remains additive
- **WHEN** a user runs `jr issue comments <url-or-key>` in any output format
- **THEN** `schemaVersion` is still `"1"` and `authorUsername` appears only as a new additional entry field
