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
