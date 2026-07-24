# Contributing to Facet

Thank you for your interest in contributing to Facet! This document outlines the process for contributing code, documentation, or bug reports.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

1. Check existing issues to avoid duplicates
2. Open a GitHub issue with the bug report template
3. Include: steps to reproduce, expected behavior, actual behavior, environment info

### Feature Requests

Open a GitHub issue with the feature request template describing the use case and proposed solution.

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes following the coding conventions
4. Add or update tests as needed
5. Run tests locally:
   ```bash
   # Backend
   cd facet-backend && go test ./...

   # Frontend
   cd facet-front && npm test
   ```
6. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation changes
   - `refactor:` code restructuring
   - `test:` adding or updating tests
   - `chore:` tooling, CI, dependencies
7. Push and open a pull request against `main`

## Coding Conventions

### Go Backend

- Follow [Effective Go](https://golang.org/doc/effective_go) and standard `gofmt` formatting
- Run `go vet ./...` before committing
- Use table-driven tests with [`testify`](https://github.com/stretchr/testify)
- Keep layered architecture: Handler → Service → Repository
- Soft deletes for users, environments, and feature flags

### React Frontend

- Maintain existing code style (no ESLint/Prettier enforced yet)
- Use Zustand stores for state management
- Use `@` alias for `src/` imports
- Place page components in `src/views/`, API modules in `src/api/`
- Cover stores and API modules with tests using vitest

## Code Review

All PRs require at least one approval. Reviewers will check:

- Correctness and test coverage
- Adherence to coding conventions
- No sensitive information (passwords, keys, IP addresses)

Thank you for contributing!
