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

### Workload Vulnerabilities

`get_workload_vulnerabilities` fetches one page for an exact application or job,
using the existing authenticated Nais API client. It includes suppressed
findings without suppression reasons. It does not query neighbouring workloads,
return manifests or secrets, contact public advisory services or analyse code.

```json
{
  "team": "my-team",
  "workload": "my-app",
  "environment": "prod-gcp",
  "workload_type": "application",
  "first": 10
}
```

Use `job` for jobs. Team, workload, environment and type are required;
`first` defaults to 10 and accepts 1–50. For the next page, pass
`page_info.end_cursor` as `after` and the first response's
`snapshot_identity` as `expected_snapshot`. Keep scope unchanged.

The structured result contains:

| Field | Contents |
|---|---|
| `workload` | ID, team, name, type and environment |
| `context` | Ingress URL/type, auth-integration type names, inbound rules (workload, team, mutual) |
| `snapshot` | Image ID/name/tag/digest, SBOM status, nullable unsuppressed total and `summary_updated_at` |
| `findings` | Node/advisory IDs, purl, severity, description, risk assessment, remediation, suppression state and details link |
| `page_info` | `has_next_page`, `end_cursor`, `requested_after` |
| `snapshot_identity` | SHA-256 fingerprint of observed workload/image/summary metadata |
| `fetched_at`, `returned_count`, `coverage_issues` | UTC fetch time, page count and data-gap codes |

Risk facts include CVSS, EPSS,
KEV and known ransomware use. `epss_score` is a probability (0–1);
`epss_percentile` is a percentage (0–100). Null scores/fix/latest versions
remain unknown. Descriptions are limited to 1000 Unicode characters with
`description_truncated: true`; use `details_link` for the full advisory.
Identifiers, risk facts and versions are not truncated.

Coverage codes are `sbom_not_ready`, `summary_unavailable` and
`scan_time_unknown`. Missing data is not a clean scan.
`summary_updated_at` is the summary's `lastUpdated`, not a scan timestamp;
recalculating a summary can change it without a new scan. The query exposes no
actual scan timestamp, so `scan_time_unknown` is always returned.
Summary counts and update time can originate from another image tag. They do
not prove current-image coverage or freshness, and the fingerprint cannot
detect a change in summary source unless selected metadata changes.
Coverage codes describe observed data gaps, not a complete inventory. Collect
every page and check advancing cursors and unique finding IDs. Summary `total` counts only
unsuppressed findings; `findings` and `returned_count` include suppressed rows.
Never compare all returned rows to this total or stop pagination at that count.
If checking aggregate coverage, compare only the unsuppressed rows after
collecting every page. A zero total can still have suppressed findings.

These semantics are established by the
[API's image-list query](https://github.com/nais/api/blob/da0d02c456dd1738d1390db3a2ddbe55a1389d64/internal/vulnerability/queries.go#L276-L335),
which requests suppressed rows, and
[v13s's summary query](https://github.com/nais/v13s/blob/0fbbf81c62c259e89423bcda09c0b2958be757e2/internal/database/queries/vulnerabilities.sql#L1-L77),
which excludes them. Source verification is not validation against a live API.

Errors include `invalid_scope`, `invalid_workload_type`, `invalid_page_size`,
`invalid_cursor`, `scope_mismatch`, `invalid_response`
and `snapshot_changed`. Authentication/API errors remain errors, never empty
findings. A changed snapshot requires a separate fresh collection, not merging
old and new pages. Do not bypass an error with a broader GraphQL query.

The fingerprint detects observable metadata changes; it is not authorization
or an atomic backend snapshot. Unreported changes cannot be detected.
Ingress/auth/network configuration does not prove reachability or running
replica identity. Explanations, recommendations and local evidence belong in
the consuming skill.

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
- Prefer `get_workload_vulnerabilities` for exact application/job vulnerability
  inventories. Follow all pages with `after` and `expected_snapshot`.
- Use pagination with reasonable page sizes (20-50 items, max 100)
- Filter queries when possible (by team, environment, name)
- Use `__typename` for union/interface types
- Include `pageInfo { hasNextPage endCursor }` for paginated results

`execute_graphql` is read-only. Suppression is the sole dedicated write flow:
prepare a proposal, then request human authorization through MCP elicitation.
Never infer permission from a vulnerability report, investigation or autopilot.
All API calls use the authenticated user's identity.
```

### Vulnerability suppression

`prepare_vulnerability_suppression` takes a `target` with exact API-returned
`findingID`, `advisoryID`, `packagePurl`, `workloadID`, `team`, `environment`,
`workload`, `workloadType` (`application` or `job`), `imageID`, `imageName`
and `imageTag` (the API value, possibly empty for digest-only images). The image
ID binds the full reference. It returns an unsaved proposal and `scopeFingerprint`.
It requires the separate Nais API suppression-preview/apply contract; it does
not fall back to the older mutation when that contract is unavailable.

Suppression applies to the **image name, versioned package purl and CVE,
including aliases**, across tags, teams, environments and future matches.
It is not workload-local. The proposal lists up to 50 observed affected
workloads with usable scans, including their team, environment and tag.
Unscanned or changed consumers can also be affected. Larger observed scopes
require manual review rather than a truncated confirmation.

`confirm_vulnerability_suppression` takes the same target and
`expected_scope_fingerprint`. After rechecking the proposal, it displays the
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
Justification and evidence are not returned by these tools or added to activity
entries. This does not grant permission to execute arbitrary GraphQL mutations.

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
