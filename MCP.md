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
- `execute_mutation` - Execute an allowlisted mutation with human confirmation
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

`execute_graphql` is read-only. `execute_mutation` accepts only explicitly
allowlisted mutations, currently `applyVulnerabilitySuppression`. Query the
preview first, then request human authorization through MCP elicitation.
Never infer permission from a vulnerability report, investigation or autopilot.
All API calls use the authenticated user's identity.
```

### Vulnerability suppression

Query `vulnerabilitySuppressionPreview(target: ...)` through `execute_graphql`.
The `target` contains exact API-returned
`findingID`, `advisoryID`, `packagePurl`, `workloadID`, `team`, `environment`,
`workload`, `workloadType` (`application` or `job`), `imageID`, `imageName`
and `imageTag` (the API value, possibly empty for digest-only images). The image
ID binds the full reference. Request the preview's `target`, `affectedWorkloads`,
`includesFutureMatches`, `warning` and `scopeFingerprint` fields to review the
unsaved proposal.
It requires the separate Nais API suppression-preview/apply contract; it does
not fall back to the older mutation when that contract is unavailable.

Suppression applies to the **image name, versioned package purl and CVE,
including aliases**, across tags, teams, environments and future matches.
It is not workload-local. The proposal lists up to 50 observed affected
workloads with usable scans, including their team, environment and tag.
Unscanned or changed consumers can also be affected. Larger observed scopes
require manual review rather than a truncated confirmation.

Call `execute_mutation` with the exact API mutation name and the reviewed proposal:

```json
{
  "mutation": "applyVulnerabilitySuppression",
  "input": {
    "target": {
      "findingID": "VUL_finding",
      "advisoryID": "CVE-2026-1234",
      "packagePurl": "pkg:npm/example@1",
      "workloadID": "APP_app",
      "team": "my-team",
      "environment": "prod-gcp",
      "workload": "my-app",
      "workloadType": "application",
      "imageID": "IMG_image",
      "imageName": "registry.example/app",
      "imageTag": "v1"
    },
    "expectedScopeFingerprint": "<scopeFingerprint from the reviewed preview>"
  }
}
```

The tool accepts a mutation name and proposal, not a GraphQL document.
Only `target` and `expectedScopeFingerprint` are accepted in this mutation's
input; `confirmed`, `decision` and unknown fields are rejected. The server
constructs the backend mutation after human authorization rather than forwarding
agent-supplied arguments.

After rechecking the proposal, it displays the
exact action, affected workloads and shared-scope warning through the client's
MCP elicitation form. The human supplies the justification, evidence reference
or explicit risk acceptance, state and authorization. Without evidence,
accepted risk must use `IN_TRIAGE`, never `FALSE_POSITIVE` or `NOT_AFFECTED`.
No model-supplied confirmation field is accepted as human consent.
Use a trusted MCP client that presents elicitation forms to the human; the
server cannot prove that a malicious client actually displayed the form.

A client without elicitation support receives an unsaved proposal; no write
is attempted. Decline, cancellation, invalid input or a changed fingerprint
also prevents the write. Confirmation expires after ten minutes.
Modern MCP uses an opaque, single-use continuation bound to the reviewed
target and scope; legacy clients use server-initiated elicitation. Pending
continuations stay in memory, expire and are not usable after a server restart.

Results distinguish `proposed`, `cancelled`, `persisted` and `indeterminate`.
Only `persisted` means the API confirmed the saved state and retained decision.
An error after submission can mean a write occurred: inspect stored state
before retrying an `indeterminate` result; never retry automatically. Other
tags' summary counts can lag behind the saved suppression. Neither the scope
fingerprint nor the observed workload list freezes backend state.

The backend validates the finding/workload/current-image relationship and the
selected team's existing suppression permission. It retains a versioned decision
in the existing suppression reason, deriving the actor from authentication.
Justification and evidence are not returned by `execute_mutation` or added to
activity entries. This does not grant permission to execute arbitrary GraphQL mutations.

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
