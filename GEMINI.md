# MCP Toolbox Context

This file (symlinked as `CLAUDE.md`, `AGENTS.md`, and `.gemini/styleguide.md`) provides context and guidelines for AI agents working on the MCP Toolbox for Databases project. It summarizes key information from `CONTRIBUTING.md` and `DEVELOPER.md`.

## Project Overview

**MCP Toolbox for Databases** is a Go-based project designed to provide Model Context Protocol (MCP) tools for various data sources and services. It allows Large Language Models (LLMs) to interact with databases and other tools safely and efficiently.

## Tech Stack

-   **Language:** Go (1.23+)
-   **Documentation:** Hugo (Extended Edition v0.146.0+)
-   **Containerization:** Docker
-   **CI/CD:** GitHub Actions, Google Cloud Build
-   **Linting:** `golangci-lint`

## Key Directories

-   `cmd/`: Application entry points.
-   `internal/sources/`: Implementations of database sources (e.g., Postgres, BigQuery).
-   `internal/tools/`: Implementations of specific tools for each source.
-   `tests/`: Integration tests.
-   `docs/`: Project documentation (Hugo site).

## Development Workflow

### Prerequisites

-   Go 1.23 or later.
-   Docker (for building container images and running some tests).
-   Access to necessary Google Cloud resources for integration testing (if applicable).

### Building and Running

1.  **Build Binary:** `go build -o toolbox`
2.  **Run Server:** `go run .` (Listens on port 5000 by default)
3.  **Run with Help:** `go run . --help`
4.  **Test Endpoint:** `curl http://127.0.0.1:5000`

### Testing

-   **Unit Tests:** `go test -race -v ./cmd/... ./internal/...`
-   **Integration Tests:**
    -   Run specific source tests: `go test -race -v ./tests/<source_dir>`
    -   Example: `go test -race -v ./tests/alloydbpg`
    -   Add new sources to `.ci/integration.cloudbuild.yaml`
-   **Linting:** `golangci-lint run --fix`

## Developing Documentation

### Prerequisites

-   Hugo (Extended Edition v0.146.0+)
-   Node.js (for `npm ci`)

### Running Local Server

1.  Navigate to `.hugo` directory: `cd .hugo`
2.  Install dependencies: `npm ci`
3.  Start server: `hugo server`

### Versioning Workflows

1.  **Deploy In-development docs**: Merges to main -> `/dev/`.
2.  **Deploy Versioned Docs**: New Release -> `/<version>/` and root.
3.  **Deploy Previous Version Docs**: Manual workflow for older versions.

## Coding Conventions

### Tool Naming

-   **Tool Name:** `snake_case` (e.g., `list_collections`, `run_query`).
    -   Do *not* include the product name (e.g., avoid `firestore_list_collections`).
-   **Tool Type:** `kebab-case` (e.g., `firestore-list-collections`).
    -   *Must* include the product name.

### Branching and Commits

-   **Branch Naming:** `feat/`, `fix/`, `docs/`, `chore/` (e.g., `feat/add-gemini-md`).
-   **Commit Messages:** [Conventional Commits](https://www.conventionalcommits.org/) format.
    -   Format: `<type>(<scope>): <description>`
    -   Example: `feat(source/postgres): add new connection option`
    -   Types: `feat`, `fix`, `docs`, `chore`, `test`, `ci`, `refactor`, `revert`, `style`.

## Adding New Features

### Adding a New Data Source

1.  Create a new directory: `internal/sources/<newdb>`.
2.  Define `Config` and `Source` structs in `internal/sources/<newdb>/<newdb>.go`.
3.  Implement `SourceConfig` interface (`SourceConfigType`, `Initialize`).
4.  Implement `Source` interface (`SourceType`).
5.  Implement `init()` to register the source.
6.  Add unit tests in `internal/sources/<newdb>/<newdb>_test.go`.

### Adding a New Tool

1.  Create a new directory: `internal/tools/<newdb>/<toolname>`.
2.  Define `Config` and `Tool` structs.
3.  Implement `ToolConfig` interface (`ToolConfigType`, `Initialize`).
4.  Implement `Tool` interface (`Invoke`, `ParseParams`, `Manifest`, `McpManifest`, `Authorized`).
5.  Implement `init()` to register the tool.
6.  Add unit tests.

### Adding Documentation

-   Add source documentation to `docs/en/resources/sources/`.
-   Add tool documentation to `docs/en/resources/tools/`.

# MCP Toolbox Style Guide

## Introduction

This style guide outlines the coding conventions and contribution standards for the Gen AI Toolbox for Databases. Adhering to these guidelines ensures consistency, readability, and maintainability across the codebase and its associated tools. This file is used by the Gemini Code Assist to perform consistent code reviews.

### Versioning

We use [Semantic Versioning](https://semver.org/), **MAJOR.MINOR.PATCH**, which increments with:
*   **MAJOR**: Breaking changes in API.
*   **MINOR**: Features added in a backward-compatible manner.
*   **PATCH**: Backward-compatible bug fixes.

## Key Principles

- **Readability:** Code should be clear and easy to understand for all contributors.
- **Consistency:** Follow established patterns in tool naming, package structure, and commit messages.
- **Testability:** All new features and bug fixes must be accompanied by comprehensive unit and integration tests.
- **Documentation:** Every new source or tool must be documented using the project's Hugo-based system.

## Pull Requests & Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for all commit messages and PR titles.

### PR Title Format

Format: `<type>[optional scope]: <description>`

- **Example:** `feat(source/postgres): add support for "new-field" field`
- **Example (Breaking Change):** `fix(tool/sql)!: change default parameter value`

### Types

| Type | Description | Version change affected |
| :--- | :--- | :--- |
| **BREAKING CHANGE** | Anything with this type or a `!` after the type/scope introduces a breaking API change. E.g. `fix!: description` or `feat!: description`. | major |
| **feat** | Adding a new feature to the codebase. | minor |
| **fix** | Fixing a bug or typo in the codebase. | patch |
| **ci** | Changes made to the continuous integration configuration files or scripts (usually the yml and other configuration files). | n/a |
| **docs** | Documentations-related PRs, including fixes on docs. | n/a |
| **chore** | Other small tasks or updates that don't fall into any of the types above. | n/a |
| **perf** | changed src code, with improvement of performance metrics. | n/a |
| **refactor** | Change src code but unlike feat, there are no tests broken and no lines lost coverage. | n/a |
| **revert** | Revert changes made in another commit. | n/a |
| **style** | updated src code, with only formatting and whitespace updates. In other words, this includes anything a code formatter or linter changes. | n/a |
| **test** | Changes made to test files. | n/a |
| **build** | Changes related to build of the projects and dependency. | n/a |

### Scopes

PRs addressing a specific source or tool should **always** add the source or tool name as scope.

The scope is formatted as `<type>/<kind>`. Common scopes include:
- `source/postgres`, `source/cloudsql-mysql`
- `tool/mssql-sql`, `tool/list-tables`
- `auth/google`

**Multiple Scopes:**
- If the PR covers multiple scopes of the same kind, separate them with a comma: `feat(source/postgres,source/alloydbpg): ...`.
- If the PR covers multiple scope types (e.g., adding a new database source and tool), disregard the scope type prefix: `feat(new-db): adding support for new-db source and tool`.

### PR Description

Every PR must include a description that follows the repository's template:

**1. Description**
A concise description of the changes (bug or feature), its impact, and a summary of the solution.

**2. PR Checklist**
- [ ] Make sure to open an issue as a bug/issue before writing your code!
- [ ] Ensure the tests and linter pass
- [ ] Code coverage does not decrease (if any source code was changed)
- [ ] Appropriate docs were updated (if necessary)
- [ ] Make sure to add `!` if this involves a breaking change

**3. Issue Reference**
Use the format: `Fixes #<issue_number> 🦕`
