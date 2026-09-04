## ADDED Requirements

### Requirement: List issue comments
The system SHALL provide `jr issue comments <url-or-key>` to retrieve comments for a Jira issue in the stable output schema.

#### Scenario: Comments are returned for issue URL
- **WHEN** a user runs `jr issue comments https://jira.example.com/browse/PROJ-123 -o json`
- **THEN** stdout includes `schemaVersion`, `issueKey`, `startAt`, `maxResults`, `total`, and a `comments` array

#### Scenario: Comments are returned for bare issue key
- **WHEN** a user runs `jr issue comments PROJ-123 --instance prod -o json`
- **THEN** the command resolves the `prod` instance and returns comments for `PROJ-123`

### Requirement: Comment fields
Each structured comment entry SHALL include stable fields for `id`, `body`, `author`, `created`, and `updated` when Jira returns them.

#### Scenario: Comment author is mapped
- **WHEN** Jira returns a comment with `author.displayName`
- **THEN** the structured output includes that display name as the comment `author`

### Requirement: Raw comment output
The system SHALL support `-o raw` for `jr issue comments` to print the verbatim Jira comments response body.

#### Scenario: Raw comments are requested
- **WHEN** a user runs `jr issue comments PROJ-123 --instance prod -o raw`
- **THEN** stdout contains Jira's raw comments JSON response

### Requirement: Existing issue get remains unchanged
The system SHALL NOT include comments in `jr issue get` as part of this change.

#### Scenario: Issue get remains focused
- **WHEN** a user runs `jr issue get PROJ-123 --instance prod -o json`
- **THEN** the issue response does not add a `comments` field
