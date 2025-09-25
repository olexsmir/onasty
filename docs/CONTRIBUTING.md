# Getting started
First of, you need have [task](https://taskfile.dev), and [mise](https://mise.jdx.dev) installed.

```bash
git clone git@github.com:olexsmir/onasty.git
cd onasty
mise use  # installs required tools for development
task api:install # installs deps to work with openapi
task web:install # installs all node_modules for the frontend gods
task docker:up # starts all docker containers
task frontend:dev # starts dev server for elm app
task run # recompiled and restart core and mailer services
```

# Code Structure Guidelines
See [Architecture](./Architecture).

# Testing
```bash
task test # runs all tests in the project
  # > backend e2e and units
  # > frontend units
task test:unit # backend unit tests
task test:e2e  # backend e2e tests
task frontend:test # frontend unit tests
```

# Coding Style & Linting
- For general editor config [editorconfig](https://editorconfig.org/) is used
- Go
  - `gofumpt`, `goimports`, `golines` are used for formatting
  - `golangci-lint` is used for linting
- Elm
  - `elm-format` is used for formatting
  - `elm-review` is used for linting

# Adding Features
1. Make sure there’s no duplicate functionality.
1. Add new endpoints to `api/openapi.yml` if needed.
1. Add unit tests or e2e tests as appropriate.
1. For new mailer events: add event struct in `internal/events` and corresponding handler in `mailer/`.

# Branching & Workflow
- Use feature branches for any changes: `feat/<description>` or `fix/<escription>`.
- Follow [Conventional commit](https://www.conventionalcommits.org/en/v1.0.0) messages.
- Open PR against `main` branch.
