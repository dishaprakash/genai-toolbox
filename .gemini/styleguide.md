# Gen AI Toolbox Style Guide

## Introduction

This style guide outlines the coding conventions and contribution standards for the Gen AI Toolbox for Databases. Adhering to these guidelines ensures consistency, readability, and maintainability across the codebase and its associated tools. This file is used by the Gemini Code Assist to perform consistent code reviews.

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

| Type                | Description                                                  | Version Bump |
| :------------------ | :----------------------------------------------------------- | :----------- |
| **BREAKING CHANGE** | Any change with `!` or `BREAKING CHANGE` in footer           | Major        |
| **feat**            | Adding a new feature                                         | Minor        |
| **fix**             | Fixing a bug or typo                                         | Patch        |
| **docs**            | Documentation-only changes                                   | None         |
| **test**            | Adding or correcting tests                                   | None         |
| **ci**              | CI/CD configuration changes                                  | None         |
| **refactor**        | Code change that neither fixes a bug nor adds a feature      | None         |
| **chore**           | Routine tasks, maintenance, or dependency updates            | None         |
| **style**           | Formatting, missing semi-colons, etc. (no code logic change) | None         |

### Scopes

Scopes should follow the format `<kind>/<name>`. Common scopes include:

- `source/postgres`, `source/bigquery`, etc.
- `tool/mssql-sql`, `tool/list-tables`, etc.
- `auth/google`

For PRs covering multiple scopes of the same kind, use commas: `feat(source/postgres,source/alloydbpg): ...`.

### PR Description

Every PR must include a description that explains:

1.  **What** changed.
2.  **Why** the change was made (impact).
3.  **How** it was solved (summary of solution).
4.  Reference to any related issues (e.g., `Fixes #123`).
