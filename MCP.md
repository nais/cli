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
- `validate_graphql` - Validate a GraphQL query without executing it

### Vulnerability Remediation
- `get_workload_vulnerability_report` - Start here when asked to fix or investigate vulnerabilities in an application or job. Resolves exact workload names, retrieves every page of Nais findings, and groups shared images across environments.

```json
{
  "team": "nais",
  "name": "oracleverk"
}
```

Optionally set `kind` to `application` or `job`, or filter with
`"environments": ["prod-gcp"]`. If both workload kinds match, the tool requires
an explicit `kind` rather than choosing a target.

The structured report contains image digests, SBOM status, summary freshness,
latest deployment metadata, package PURLs with installed versions, severity,
operational risk signals, fix/latest versions, advisory links and existing
suppressions. Shared images are grouped by image name and digest, falling back
to name and tag with a warning when the digest is unavailable. Finding fetches
fail if the selected workload changes image during pagination.

Compare deployed versions with the current source and build configuration before
upgrading: source may already be fixed and only need a rebuild/deploy. Latest
deployment metadata is not proof of the scanned image's source commit. A complete
report means all pages were retrieved, not that scanning is current or the image
is vulnerability-free. Missing SBOMs, unavailable summaries and stale summaries
produce explicit warnings.

This tool only reads data. It does not change source, build, deploy or suppress
findings. Require evidence and explicit approval for suppression, and do not claim
deployed remediation until the intended new image has current scan results.

### Schema Exploration
- `schema_list_types` - List all types in the API schema
- `schema_get_type` - Get details about a specific type
- `schema_list_queries` - List all available query operations
- `schema_list_mutations` - List all mutation operations (read-only server)
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
2. For vulnerability remediation or triage, call `get_workload_vulnerability_report` with the team/name from the Nais manifest; no schema exploration is needed
3. Otherwise use schema exploration tools (`schema_list_queries`, `schema_get_type`) to discover available data
4. Construct GraphQL queries based on the schema
5. Execute queries with `execute_graphql`

**Query Guidelines:**
- Use pagination with reasonable page sizes (20-50 items, max 100)
- Filter queries when possible (by team, environment, name)
- Use `__typename` for union/interface types
- Include `pageInfo { hasNextPage endCursor }` for paginated results

All operations are read-only and use the user's authenticated identity.
```

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
