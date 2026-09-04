# Issue Assign Delta

## ADDED Requirements

### Requirement: Assign issue to a username
The system SHALL provide `jr issue assign <url-or-key> <username>` to assign a
Jira issue to the user identified by an exact username.

#### Scenario: Assign by issue URL
- **WHEN** a user runs `jr issue assign https://jira.example.com/browse/PROJ-123 jdoe`
- **THEN** the CLI sends `PUT /rest/api/2/issue/PROJ-123/assignee` with body `{"name":"jdoe"}` and renders a structured confirmation

#### Scenario: Assign bare issue key with instance
- **WHEN** a user runs `jr issue assign PROJ-123 jdoe --instance prod`
- **THEN** the command resolves the `prod` instance and performs the same assignment

#### Scenario: Blank username is rejected
- **WHEN** a user runs `jr issue assign <url-or-key>` with an empty or whitespace-only username
- **THEN** the CLI fails without any HTTP request, with a one-sentence error and a next step

### Requirement: Assign confirmation schema
The system SHALL render a successful assignment as a structured response with
`schemaVersion: "1"`, `issueKey`, `assignee` (the username exactly as
requested), and `url`, in the active output format. The `assignee` value SHALL
be the requested username echoed back, not a value re-fetched from Jira.

#### Scenario: Confirmation includes assignee username
- **WHEN** Jira responds `204 No Content` to the assignee update
- **THEN** structured output includes `schemaVersion`, `issueKey`, `assignee: jdoe`, and the browse `url`

### Requirement: Unassignable user error is actionable
The system SHALL report a rejected assignment as a one-sentence error with a
suggested next step, and SHALL NOT make any additional API call to enumerate
assignable users.

#### Scenario: Jira rejects the username
- **WHEN** Jira responds with an error because the user is not assignable
- **THEN** stderr contains a one-sentence error plus a next step directing the user to the exact username in `jr issue get` / `jr issue comments` output

### Requirement: Exact usernames only
The system SHALL send the username argument verbatim as the assignee `name`
and SHALL NOT perform any user lookup, search, or fuzzy matching to resolve
display names or email addresses.

#### Scenario: Argument is sent verbatim
- **WHEN** a user passes any string as the assign target
- **THEN** the CLI sends it as-is in the assignee request and makes no other user-resolution API call
