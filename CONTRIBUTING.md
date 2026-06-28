<!-- gambits-managed -->
# Contributing to aws-infra-scaler

Thanks for taking the time to contribute! Please follow these guidelines to keep the codebase consistent.

## Code Style

- **Indentation:** Use 2 spaces throughout — no tabs, no 4-space indents. This applies to all file types (JS, TS, JSON, YAML, etc.).

## Commit Messages

This repo follows the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <short summary>
```

### Types

| Type       | When to use                                      |
|------------|--------------------------------------------------|
| `feat`     | A new feature                                    |
| `fix`      | A bug fix                                        |
| `docs`     | Documentation changes only                       |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test`     | Adding or updating tests                         |
| `chore`    | Maintenance tasks (deps, CI, tooling)            |
| `perf`     | Performance improvement                          |

### Examples

```
feat(ecs): add support for scaling ECS services by CPU threshold
fix(rds): handle missing instance identifier gracefully
docs: add CONTRIBUTING.md
chore(deps): bump aws-sdk to v3.x
```

### Rules

- Use the **imperative mood** in the summary line ("add", not "added" or "adds").
- Keep the summary line under 72 characters.
- Reference relevant GitHub issues in the body where applicable (e.g. `Closes #42`).

## Pull Requests

- Open PRs against the default branch.
- Ensure your branch is up to date before requesting review.
- One logical change per PR — keep them small and focused.
