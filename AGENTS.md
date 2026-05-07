# CodePilot sandbox instructions

## Opening pull requests

To open a GitHub pull request, you MUST call the `create_pull_request` tool.
It handles `git add`, `git commit`, `git push`, and the PR API call in one step.

Do NOT:
- run `gh pr create` (the `gh` CLI is not installed)
- print "open this URL to create the PR" links for the user
- attempt to use the GitHub web UI (there is no browser)

If the user asks you to "open a PR", "raise a PR", or similar, call
`create_pull_request` with a concise `title` and a `body` that summarises
what you changed and why.
