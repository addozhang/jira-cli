### Requirement: Get issue
The system SHALL provide `jr issue get <url-or-key>` to retrieve a Jira issue in the stable output schema.

#### Scenario: Issue is found
- **WHEN** a user runs `jr issue get https://jira.example.com/browse/PROJ-123`
- **THEN** stdout includes schema version, issue key, summary, status, type, project, assignee, reporter, description, labels, and updated timestamp when available

### Requirement: Add comment
The system SHALL provide `jr issue comment <url-or-key>` to add a comment body as verbatim Jira Server wiki markup.

#### Scenario: Comment from stdin
- **WHEN** a user runs `jr issue comment PROJ-123 --body - --instance prod`
- **THEN** the command reads stdin and sends it as the comment body without conversion

### Requirement: Write commands are explicit
The system SHALL require comment writes to receive explicit command-line intent and SHALL NOT perform interactive prompts for the initial issue workflow.

#### Scenario: Missing comment body
- **WHEN** a user runs `jr issue comment PROJ-123 --instance prod` without `--body`
- **THEN** the command fails before making an HTTP request and asks for `--body`
