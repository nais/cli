# AGENTS.md - Nais CLI

Guidelines for AI agents working on this project.

## Nais MCP Server

See `MCP.md` for documentation on the Nais MCP server, which lets AI agents interact with the Nais platform through GraphQL.

## Project Structure

- **Go project** using naistrix (which wraps Cobra) for CLI commands
- Commands in `internal/`, organized by domain
- GraphQL client generated with genqlient
- Schema in `schema.graphql`

## Common Commands

| Task | Command |
|---------|----------|
| Build | `mise run build` |
| Run tests | `mise run test` |
| Format code | `mise run fmt` |
| Run all checks | `mise run check` |
| Static analysis | `mise run check:staticcheck` |
| Generate GraphQL | `mise run generate:graphql` |
| Generate documentation | `mise run generate:docs` |
| Serve documentation locally | `mise run docs` |
| Update schema (local API) | `mise run update:graphql-schema:local` |
| Update schema (live API) | `mise run update:graphql-schema:live` |
| Run locally against the local API | `mise run local` |

## Running the CLI Locally

```bash
mise run build
./bin/nais --version
./bin/nais <command>
```

## Workflow

- **Branches**: Make changes and commits on a descriptive feature, fix, or docs branch, not directly on `main` or other default/protected branches. Check the current branch before editing code or committing, and create an appropriate branch if needed. Reuse an existing task branch when appropriate. Commit directly to a default/protected branch only when the user explicitly requests it.

1. **Before changing code**: Read the relevant files to understand existing patterns
2. **After changing GraphQL queries**: Run `mise run generate:graphql`
3. **When the API schema changes**: Run `mise run update:graphql-schema:local`
4. **After all changes**: Run `mise run test` and `mise run fmt`

## GraphQL

- The schema is in `schema.graphql` (fetched from nais-api)
- The client is generated with genqlient (configured in `genqlient.yaml`)
- Queries are defined in Go files using `# @genqlient` comments

## Key Conventions

- **Command structure**: Commands are generally grouped by domain under `internal/<domain>/command/`, with individual commands in separate files
- **Formatting**: `gofumpt`
- **Commit messages**: Conventional Commits (see `script/semantic-commit-hook.sh`)

## Documentation

- The CLI manual is generated with naistrix and built with Jekyll using the Just the Docs theme. See `docs/README.md` for local setup.
- Edit command titles, descriptions, flag help, and examples in the Go source, not in generated Markdown.
- Register defaults key documentation with `naistrix.ApplicationWithDefaultsDocumentation` in `internal/application/application.go`. Keep `Title` short (aim for 80 visible characters or fewer) and use `Description` for detailed guidance and examples, using `heredoc.Doc` for multiline prose.
- Run `mise run generate:docs` after changing documentation in Go. Command pages are moved into the generated `docs/_commands/` collection, while `docs/configuration.md` remains a regular page. Both are ignored by Git.
- `mise run docs` generates the documentation once and starts Jekyll with browser live reload. Go source changes require rerunning `mise run generate:docs`; there is no Go source watcher.

## Postgres CLI completion contract

- Complete Postgres names from `Team.postgreses` (`GetTeamPostgreses`), including instances without branches. Branches are internal API details: do not expose branch commands, flags, completions or output fields. Personal access always targets the observed active branch.
- With no environment, resolve from that Postgres's team environments: auto-select one, prompt for multiple, or require `-e` when noninteractive. Propagate API/auth errors in completion help (including login guidance).
- Top-level `postgres create <postgres>` must not suggest existing names; update/delete may. `--version` suggests 18; update `--high-availability` suggests true/false.
- Global `--environment` completion lists all environments; do not shadow it or add command-specific global flag completion. Deploy the API schema/mutations before the CLI; completion and listing require the API's `Team.postgreses` contract.

## Local Development

```bash
mise install
mise run build
./bin/nais --version
```

For shell completion:
```bash
source <(./bin/nais completion zsh)
```
