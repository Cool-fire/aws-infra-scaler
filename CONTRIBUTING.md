# Contributing to AWS Infra Scaler

Thanks for your interest in contributing! This document outlines how to propose changes, report issues, and get your work merged.

## Getting Started

1. Fork the repository and clone your fork.
2. Ensure you have Go 1.19+ installed.
3. Install dependencies:
   ```
   go mod download
   ```
4. Build the binary:
   ```
   make
   ```
5. Run the tool against a test configuration to confirm your setup works.

## Development Workflow

1. Create a feature branch off `main`:
   ```
   git checkout -b my-change
   ```
2. Make your changes, keeping commits focused and logically grouped.
3. Run `go build ./...` and `go vet ./...` before pushing.
4. Run `gofmt -s -w .` to format your code.
5. Push your branch and open a pull request against `main`.

## Reporting Issues

When filing an issue, please include:

- A clear description of the problem or feature request.
- Steps to reproduce (for bugs), including the relevant section of your `config.yaml` with any sensitive values redacted.
- The CLI flags used and the full error output.
- Your Go version, OS, and AWS region.

## Pull Requests

- Keep PRs small and focused — one change per PR when possible.
- Write a descriptive PR title and summary explaining the *why*, not just the *what*.
- Reference the issue your PR addresses (e.g., `Fixes #42`).
- Update the `README.md` if you change user-facing behavior, flags, or configuration.
- Ensure the build passes and the tool works end-to-end with a real or mocked AWS setup.

## Code Guidelines

- Follow standard Go conventions: `gofmt`, idiomatic error handling, exported identifiers documented.
- Keep service-specific logic inside `pkg/service/` — one file per AWS service.
- Configuration types belong in `pkg/config/config_types.go`; parsing in `pkg/config/config.go`.
- Return errors up to the caller rather than logging and swallowing them. Region-level error accumulation is handled in `pkg/scaler.go`.
- Avoid introducing new top-level dependencies without discussion.

## Adding Support for a New AWS Service

1. Add a new config struct in `pkg/config/config_types.go` and wire it into the YAML decoder in `pkg/config/config.go`.
2. Create `pkg/service/<service>.go` implementing scale-up and scale-down logic using `aws-sdk-go-v2`.
3. Register the service in `pkg/service/service.go` so the scaler can dispatch to it.
4. Document the new service and its configuration schema in `README.md`.
5. Include an example block in `config.yaml`.

## Commit Messages

- Use the imperative mood ("Add Kinesis retry logic", not "Added" or "Adds").
- Keep the subject line under 72 characters.
- Add a body when the change needs context — explain motivation and any non-obvious tradeoffs.

## Code of Conduct

Be respectful and constructive in all interactions. Assume good intent, and give feedback on the code rather than the contributor.
