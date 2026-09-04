### Requirement: Raw JQL search
The system SHALL provide `jr search --jql <query>` to execute raw Jira Query Language against a selected instance.

#### Scenario: Raw JQL executes
- **WHEN** a user runs `jr search --jql "project = ENG" --instance prod`
- **THEN** the command executes the supplied JQL query without rewriting it

### Requirement: Search instance selection
The system SHALL require `--instance URL|alias` for search.

#### Scenario: Search with alias instance
- **WHEN** a user runs `jr search --jql "project = ENG" --instance prod`
- **THEN** the command executes the query against the instance referenced by alias `prod`

### Requirement: Bounded pagination
The system SHALL return a single bounded page of search results.

#### Scenario: Search maps to bounded Jira request
- **WHEN** a user runs `jr search --jql "project = ENG" --instance prod`
- **THEN** the Jira request uses a bounded `maxResults` value

### Requirement: Search output schema
The system SHALL render search results with schema version, query, pagination metadata, and issue summaries.

#### Scenario: Search returns issues
- **WHEN** Jira returns matching issues
- **THEN** stdout includes `schemaVersion`, the executed JQL, pagination metadata, and a list of issue summaries
