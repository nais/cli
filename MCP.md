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
- `get_workload_vulnerability_report` - Start here when asked to fix or investigate vulnerabilities. Returns a compact package-grouped overview for an application or job.
- `get_workload_vulnerability_status` - After deploy, check the intended new digest and scan readiness without fetching individual findings.

```json
{
  "team": "nais",
  "name": "oracleverk"
}
```

Optionally set `kind` to `application` or `job`, or filter with
`"environments": ["prod-gcp"]`. If both workload kinds match, the tool requires
an explicit `kind` rather than choosing a target.

The report retrieves every finding page, but by default returns only an overview
per installed package PURL: finding counts by severity and operational priority,
distinct reported fix/latest versions, counts without known fixes and counts with
suppressions. Finding counts are not unique CVE counts. Reported fix versions may
span release branches; the tool does not pick an upgrade by string ordering.

Set `"include_details": true` when individual advisories, risk signals or suppression
reasons are needed. This adds full findings alongside the same package overview.

Image digests, SBOM status, summary freshness and latest deployment metadata are
included in either mode. Shared images are grouped by image name and digest,
falling back to name and tag with a warning when the digest is unavailable.
Finding fetches fail if the selected workload changes image during pagination.

Compare deployed versions with the current source and build configuration before
upgrading: source may already be fixed and only need a rebuild/deploy. Latest
deployment metadata is not proof of the scanned image's source commit. A complete
report means all pages were retrieved, not that scanning is current or the image
is vulnerability-free. Missing SBOMs, unavailable summaries and stale summaries
produce explicit warnings.

For post-deploy verification, call `get_workload_vulnerability_status`:

```json
{
  "team": "nais",
  "name": "oracleverk",
  "environments": ["dev-gcp", "prod-gcp"],
  "expected_digest": "sha256:<digest-from-the-new-build>"
}
```

Use the digest of the intended **new** image from the build/deploy output, not the
pre-fix report. The tool returns `digest_unavailable`, `not_deployed`, `scan_pending`,
`scan_unavailable` or `scan_ready` per group. It verifies all explicitly requested
environments exist, and checks each workload's scan metadata even when images are
shared. Use separate calls when environments should have different digests.

`verified: true` means every selected workload uses the expected digest, its SBOM
is READY and its summary has an update timestamp and no older-image fallback.
It does **not** mean zero vulnerabilities, no suppressions or verified source
provenance. Inspect the returned summary counts separately. The API's summary
update timestamp is not an authoritative scan timestamp. This call does not wait
or poll; repeat it with a reasonable delay if the rollout or scan is pending.

Both tools only read Nais data. They do not change source, build, deploy or suppress
findings. Require evidence and explicit approval for suppression.

#### Current API gaps

The API does not expose advisory aliases/withdrawal or a verified image-to-source
build relationship. The MCP does not guess these values or fetch external
advisories automatically. In particular, an unknown fix version is not evidence
that a finding is unfixable or should be suppressed. Latest deployment metadata
remains context, not image build provenance.

#### Fresh-session acceptance check

With the local MCP binary configured and an application repository open, start a
fresh agent session and ask only: "Fix vulnerabilities in this app."
Check that it calls the report tool, resolves application vs job and environments,
compares deployed packages with the checkout before editing, and requests detail
only as needed. After an approved deploy, it should check the intended new digest
using the status tool and distinguish missing/stale scan data from zero findings.
Unit tests and direct tool calls do not establish this agent-routing behavior.

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
   - Start with the compact overview; request `include_details: true` only as needed
   - After deploy, use `get_workload_vulnerability_status` with the intended new digest
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
