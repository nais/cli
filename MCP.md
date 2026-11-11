# Nais MCP Server

The Nais MCP (Model Context Protocol) server allows LLMs and AI assistants to interact with the Nais platform. It provides dynamic access to the Nais GraphQL API through schema exploration and query execution tools.

## Quick Start

### Installation

Ensure you're authenticated with the Nais CLI:

```bash
nais login -n
```

### Configuration

Add to your MCP settings file:

**GitHub Copilot CLI** (`~/.mcp/config.json`):
```json
{
  "mcpServers": {
    "nais": {
      "command": "nais",
      "args": ["alpha", "mcp", "serve"],
			"tools": ["*"]
    }
  }
}
```

**Zed** (`~/.config/zed/settings.json`):
```json
{
  "context_servers": {
    "nais": {
      "enabled": true,
      "command": "nais",
      "args": ["alpha", "mcp", "serve"]
    }
  }
}
```

**VS Code** (with Cline extension):

1. Open the command palette
2. Select "MCP: Add Server..."
3. Select "Command (stdio)"
4. Insert `nais` in the command input and press Enter
5. Insert `alpha mcp serve` in the args input and press Enter
6. When prompted for a name, insert `nais`
7. Select if you want to add it as a Global or Workspace MCP server

**IntelliJ IDEA** (with GitHub Copilot):

See [GitHub Copilot MCP documentation](https://docs.github.com/en/copilot/how-tos/provide-context/use-mcp/extend-copilot-chat-with-mcp?tool=jetbrains) for setup instructions.

Local server configuration:
```json
{
  "servers": {
    "nais": {
      "command": "nais",
      "args": [
        "alpha",
        "mcp",
        "serve"
      ]
    }
  }
}
```

## Available Tools

### Context & Execution
- `get_nais_context` - Get current user, teams, and console URL patterns
- `execute_graphql` - Execute GraphQL queries against the Nais API
- `execute_mutation` - Execute an allowlisted mutation with human confirmation (stdio only)
- `validate_graphql` - Validate a GraphQL query without executing it

### Schema Exploration
- `schema_list_types` - List all types in the API schema
- `schema_get_type` - Get details about a specific type
- `schema_list_queries` - List all available query operations
- `schema_list_mutations` - Discover mutation definitions; does not execute them
- `schema_get_field` - Get details about a specific field
- `schema_get_enum` - Get enum values and descriptions
- `schema_search` - Search the schema by name or description
- `schema_get_implementors` - Get types implementing an interface
- `schema_get_union_types` - Get member types of a union

## Recommended Agent Prompt

Add this to your `AGENTS.md` or system prompt to help the LLM use the Nais MCP effectively:

```markdown
You have access to the Nais MCP server for interacting with the Nais platform.

**Initial Setup:**
1. Always start with `get_nais_context` to understand the user, their teams, and available console URLs
2. Use schema exploration tools (`schema_list_queries`, `schema_get_type`) to discover available data
3. Construct GraphQL queries based on the schema
4. Execute queries with `execute_graphql`

**Query Guidelines:**
- Use pagination with reasonable page sizes (20-50 items, max 100)
- Filter queries when possible (by team, environment, name)
- Use `__typename` for union/interface types
- Include `pageInfo { hasNextPage endCursor }` for paginated results

`execute_graphql` is read-only. `execute_mutation` is available only over stdio
and accepts only explicitly allowlisted mutations, currently
`updateImageVulnerability`. Read the finding first, then request human
authorization through MCP elicitation.
Never infer permission from a vulnerability report, investigation or autopilot.
All API calls use the authenticated user's identity.
```

### Vulnerability suppression

Mutation execution is restricted to stdio. HTTP/SSE transports expose read-only
tools because they do not authenticate MCP clients. Elicitation confirms an
action with the connected client; it does not authenticate that client.

Suppression applies to the **image name, versioned package purl and CVE,
including aliases**, across tags, teams, environments and future matches.
It is not workload-local. The confirmation warns about this shared scope but
does not enumerate all affected workloads.

Read the finding through `execute_graphql`, then pass its exact `id`:

```json
{
  "mutation": "updateImageVulnerability",
  "input": {
    "vulnerabilityID": "<exact finding id returned by the API>"
  }
}
```

Only `vulnerabilityID` is accepted in `input`. The human supplies a custom,
nonblank reason (up to 4096 bytes), state and authorization in the elicitation
form. The reason is saved unchanged; it can include an explanation or evidence
references. No particular state is forced:

| State | Meaning |
|-------|---------|
| `IN_TRIAGE` | The finding is still being assessed. |
| `FALSE_POSITIVE` | The finding is incorrect. |
| `NOT_AFFECTED` | The vulnerability does not apply. |
| `RESOLVED` | The finding has been resolved. |

The server constructs `UpdateImageVulnerabilityInput` with `suppress: true`
only after confirmation. Agent-supplied reason, state, actor, confirmation
fields and GraphQL documents are rejected.

Use a trusted client that displays the form; the server cannot prove it did.
Without elicitation support, the tool returns an unsaved proposal.
Decline, cancellation, invalid input or changes to the displayed finding prevent
the write.
Confirmation expires after ten minutes. Modern MCP uses a single-use,
in-memory continuation bound to the finding and proposal; restarting the server
invalidates it. Legacy clients use server-initiated elicitation.

| Result | Meaning |
|--------|---------|
| `proposed` | Not saved. |
| `cancelled` | Not authorized; no write attempted. |
| `persisted` | The API returned the requested suppression state and reason. |
| `indeterminate` | A write may have happened. Inspect stored state before retrying; never retry automatically. |

The existing API enforces suppression permissions, derives the actor from the
configured Nais authentication and records an activity entry with the actor and
previous/new suppression, including the reason. Do not enter secrets or personal
information. The mutation tool does not repeat the reason in its output.

The finding is reread before writing. This is not an atomic scope check:
workloads or shared scope can change between confirmation and mutation.
Other tags' summary counts may lag. No new API operations are required.

## Command Reference

```bash
nais alpha mcp serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--transport`, `-t` | `stdio` | Transport: `stdio`, `http`, or `sse` |
| `--listen`, `-l` | `127.0.0.1:8080` | Listen address (for http/sse) |
| `--rate-limit`, `-r` | `10` | Max requests per minute (0 = unlimited) |
| `--log-file` | - | Write logs to file instead of stderr |

## Resources

The server exposes these resources:

- `nais://schema` - Complete Nais GraphQL API schema
- `nais://api-best-practices` - API usage guidelines (pagination, optimization, rate limiting)
