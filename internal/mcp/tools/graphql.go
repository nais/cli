package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/naisapi"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

const maxQueryDepth = 15

// Nais API guidance for LLMs - provides context about the API structure and common patterns
const naisAPIGuidance = `
## Nais API Guidance

The Nais API is a GraphQL API for managing applications and jobs on the Nais platform.

### Key Concepts

- **Team**: The primary organizational unit. All resources belong to a team.
- **Application**: A long-running workload (deployment) managed by Nais.
- **Job**: A scheduled or one-off workload (CronJob/Job) managed by Nais.
- **Environment**: A Kubernetes cluster/namespace where workloads run (e.g., "dev", "prod").
- **Workload**: A union type representing either an Application or Job.

### Common Query Patterns

1. **Get current user and their teams**:
   ` + "```graphql" + `
   query { me { ... on User { email teams(first: 50) { nodes { team { slug } } } } } }
   ` + "```" + `

2. **Get team details**:
   ` + "```graphql" + `
   query($slug: Slug!) { team(slug: $slug) { slug purpose slackChannel } }
   ` + "```" + `

3. **List applications for a team**:
   ` + "```graphql" + `
   query($slug: Slug!) {
     team(slug: $slug) {
       applications(first: 50) {
         nodes { name state teamEnvironment { environment { name } } }
       }
     }
   }
   ` + "```" + `

4. **Get application details with instances**:
   ` + "```graphql" + `
   query($slug: Slug!, $name: String!, $env: [String!]) {
     team(slug: $slug) {
       applications(filter: { name: $name, environments: $env }, first: 1) {
         nodes {
           name state
           instances { nodes { name restarts status { state message } } }
           image { name tag }
         }
       }
     }
   }
   ` + "```" + `

5. **List jobs for a team**:
   ` + "```graphql" + `
   query($slug: Slug!) {
     team(slug: $slug) {
       jobs(first: 50) {
         nodes { name state schedule { expression } teamEnvironment { environment { name } } }
       }
     }
   }
   ` + "```" + `

6. **Get vulnerabilities for a workload**:
   ` + "```graphql" + `
   query($slug: Slug!, $name: String!, $env: [String!]) {
     team(slug: $slug) {
       applications(filter: { name: $name, environments: $env }, first: 1) {
         nodes {
           image {
             vulnerabilitySummary { critical high medium low }
             vulnerabilities(first: 20) { nodes { identifier severity package } }
           }
         }
       }
     }
   }
   ` + "```" + `

7. **Get cost information**:
   ` + "```graphql" + `
   query($slug: Slug!) {
     team(slug: $slug) {
       cost { monthlySummary { sum } }
       environments {
         environment { name }
         cost { daily(from: "2024-01-01", to: "2024-01-31") { sum } }
       }
     }
   }
   ` + "```" + `

8. **Search across resources**:
   ` + "```graphql" + `
   query($query: String!) {
     search(filter: { query: $query }, first: 20) {
       nodes {
         __typename
         ... on Application { name team { slug } }
         ... on Job { name team { slug } }
         ... on Team { slug }
       }
     }
   }
   ` + "```" + `

9. **Get alerts for a team**:
   ` + "```graphql" + `
   query($slug: Slug!) {
     team(slug: $slug) {
       alerts(first: 50) {
         nodes { name state teamEnvironment { environment { name } } }
       }
     }
   }
   ` + "```" + `

10. **Get deployments**:
    ` + "```graphql" + `
    query($slug: Slug!) {
      team(slug: $slug) {
        deployments(first: 20) {
          nodes { createdAt repository commitSha statuses { nodes { state } } }
        }
      }
    }
    ` + "```" + `

### Important Types

- **Slug**: A string identifier for teams (e.g., "my-team")
- **Cursor**: Used for pagination (pass to "after" argument)
- **Date**: Format "YYYY-MM-DD" for cost queries
- **ApplicationState**: RUNNING, NOT_RUNNING, UNKNOWN
- **JobState**: RUNNING, NOT_RUNNING, UNKNOWN
- **Severity**: CRITICAL, HIGH, MEDIUM, LOW, UNASSIGNED (for issues/vulnerabilities)

### Pagination

Most list fields support pagination with:
- ` + "`first: Int`" + ` - Number of items to fetch
- ` + "`after: Cursor`" + ` - Cursor from previous page
- ` + "`pageInfo { hasNextPage endCursor totalCount }`" + ` - Pagination info

### Filtering

Many fields support filters:
- ` + "`filter: { name: String, environments: [String!] }`" + ` - For applications/jobs
- ` + "`filter: { severity: Severity }`" + ` - For issues

### Tips

1. DO NOT query secret-related types/fields (Secret, SecretValue, etc.)
2. Always use ` + "`__typename`" + ` when querying union/interface types (Workload, Issue, etc.)
3. Use fragment spreads for type-specific fields: ` + "`... on Application { ingresses { url } }`" + `
4. Start with schema exploration to discover available fields
5. Use pagination for large result sets (default to first: 50)

### Nais Console URLs

When providing links to the user, use the console URL patterns provided by the ` + "`get_nais_context`" + ` tool.
Call ` + "`get_nais_context`" + ` to get the base URL and all available URL patterns with placeholders.
Replace the placeholders (e.g., ` + "`{team}`" + `, ` + "`{env}`" + `, ` + "`{app}`" + `) with actual values from query results.

**Note**: Do NOT invent or guess URLs. Only use the URL patterns from ` + "`get_nais_context`" + ` with actual data from query results.
`

func registerGraphQLTools(s *server.MCPServer, ctx *toolContext) {
	// get_nais_context tool - provides essential context for working with Nais
	getNaisContextTool := mcp.NewTool("get_nais_context",
		mcp.WithDescription("Get the current Nais context including authenticated user, their teams, and console URL. Call this first to understand what the user has access to and to get the correct console URL for links."),
		mcp.WithInputSchema[GetNaisContextInput](),
		mcp.WithOutputSchema[GetNaisContextOutput](),
	)
	s.AddTool(getNaisContextTool, mcp.NewStructuredToolHandler(ctx.handleGetNaisContext))

	// execute_graphql tool - dynamic output, so we use NewTypedToolHandler
	executeGraphQLTool := mcp.NewTool("execute_graphql",
		mcp.WithDescription(`Execute a GraphQL query against the Nais API.

IMPORTANT: Before using this tool, use the schema exploration tools (schema_list_queries, schema_get_type, schema_get_field) to understand the available types and fields.

This tool only supports queries (read operations). Mutations are not allowed.

`+naisAPIGuidance),
		mcp.WithInputSchema[ExecuteGraphQLInput](),
		// Note: Output is dynamic JSON from the GraphQL API, so we don't use WithOutputSchema here
	)
	s.AddTool(executeGraphQLTool, mcp.NewTypedToolHandler(ctx.handleExecuteGraphQL))

	// validate_graphql tool
	validateGraphQLTool := mcp.NewTool("validate_graphql",
		mcp.WithDescription("Validate a GraphQL query against the schema without executing it. Use this to check if your query is valid before executing."),
		mcp.WithInputSchema[ValidateGraphQLInput](),
		mcp.WithOutputSchema[ValidateGraphQLOutput](),
	)
	s.AddTool(validateGraphQLTool, mcp.NewStructuredToolHandler(ctx.handleValidateGraphQL))
}

func (t *toolContext) handleGetNaisContext(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args GetNaisContextInput,
) (GetNaisContextOutput, error) {
	if !t.rateLimiter.Allow() {
		return GetNaisContextOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	t.logger.Debug("Executing get_nais_context tool")

	// Get current user
	user, err := t.client.GetCurrentUser(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get current user", "error", err)
		return GetNaisContextOutput{}, fmt.Errorf("failed to get current user: %w", err)
	}

	// Get user's teams
	teams, err := t.client.GetUserTeams(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get user teams", "error", err)
		return GetNaisContextOutput{}, fmt.Errorf("failed to get user teams: %w", err)
	}

	// Build teams list
	teamsList := make([]NaisTeamInfo, 0, len(teams))
	for _, team := range teams {
		teamsList = append(teamsList, NaisTeamInfo{
			Slug:    team.Team.Slug,
			Purpose: team.Team.Purpose,
			Role:    string(team.Role),
		})
	}

	// Get console URL
	consoleBaseURL := t.getConsoleBaseURL(reqCtx)

	return GetNaisContextOutput{
		User: NaisUserInfo{
			Name: user.Name,
		},
		Teams:          teamsList,
		ConsoleBaseURL: consoleBaseURL,
		ConsoleURLPatterns: map[string]string{
			"team":                        "/team/{team}",
			"team_applications":           "/team/{team}/applications",
			"team_jobs":                   "/team/{team}/jobs",
			"team_alerts":                 "/team/{team}/alerts",
			"team_issues":                 "/team/{team}/issues",
			"team_vulnerabilities":        "/team/{team}/vulnerabilities",
			"team_cost":                   "/team/{team}/cost",
			"team_deployments":            "/team/{team}/deploy",
			"application":                 "/team/{team}/{env}/app/{app}",
			"application_cost":            "/team/{team}/{env}/app/{app}/cost",
			"application_deploys":         "/team/{team}/{env}/app/{app}/deploys",
			"application_logs":            "/team/{team}/{env}/app/{app}/logs",
			"application_manifest":        "/team/{team}/{env}/app/{app}/manifest",
			"application_vulnerabilities": "/team/{team}/{env}/app/{app}/vulnerabilities",
			"job":                         "/team/{team}/{env}/job/{job}",
			"job_cost":                    "/team/{team}/{env}/job/{job}/cost",
			"job_deploys":                 "/team/{team}/{env}/job/{job}/deploys",
			"job_logs":                    "/team/{team}/{env}/job/{job}/logs",
			"job_manifest":                "/team/{team}/{env}/job/{job}/manifest",
			"job_vulnerabilities":         "/team/{team}/{env}/job/{job}/vulnerabilities",
			"postgres":                    "/team/{team}/{env}/postgres/{name}",
			"opensearch":                  "/team/{team}/{env}/opensearch/{name}",
			"valkey":                      "/team/{team}/{env}/valkey/{name}",
			"bucket":                      "/team/{team}/{env}/bucket/{name}",
			"bigquery":                    "/team/{team}/{env}/bigquery/{name}",
			"kafka":                       "/team/{team}/{env}/kafka/{name}",
		},
	}, nil
}

func (t *toolContext) handleExecuteGraphQL(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args ExecuteGraphQLInput,
) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	variablesStr := args.Variables
	if variablesStr == "" {
		variablesStr = "{}"
	}

	t.logger.Debug("Executing GraphQL query", "query_length", len(args.Query), "has_variables", variablesStr != "{}")

	// Parse variables
	var variables map[string]any
	if err := json.Unmarshal([]byte(variablesStr), &variables); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid variables JSON: %v", err)), nil
	}

	// Validate the query
	validationResult, err := t.validateGraphQLQuery(reqCtx, args.Query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to validate query: %v", err)), nil
	}
	if !validationResult.valid {
		return mcp.NewToolResultError(fmt.Sprintf("invalid query: %s", validationResult.error)), nil
	}

	// Execute the query
	gqlClient, err := naisapi.GraphqlClient(reqCtx)
	if err != nil {
		t.logger.Error("Failed to create GraphQL client", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to create GraphQL client: %v", err)), nil
	}

	// Create the request
	gqlReq := &graphql.Request{
		Query:     args.Query,
		Variables: variables,
	}

	// Execute
	var response map[string]any
	resp := &graphql.Response{Data: &response}

	err = gqlClient.MakeRequest(reqCtx, gqlReq, resp)
	if err != nil {
		t.logger.Error("GraphQL query failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("query execution failed: %v", err)), nil
	}

	// Check for GraphQL errors in response
	if len(resp.Errors) > 0 {
		var errMsgs []string
		for _, e := range resp.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		return mcp.NewToolResultError(fmt.Sprintf("GraphQL errors: %s", strings.Join(errMsgs, "; "))), nil
	}

	// Marshal the response
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (t *toolContext) handleValidateGraphQL(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args ValidateGraphQLInput,
) (ValidateGraphQLOutput, error) {
	if !t.rateLimiter.Allow() {
		return ValidateGraphQLOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	result, err := t.validateGraphQLQuery(reqCtx, args.Query)
	if err != nil {
		return ValidateGraphQLOutput{}, fmt.Errorf("validation failed: %w", err)
	}

	if result.valid {
		return ValidateGraphQLOutput{
			Valid:         true,
			OperationType: result.operationType,
			OperationName: result.operationName,
			Depth:         result.depth,
		}, nil
	}

	return ValidateGraphQLOutput{
		Valid: false,
		Error: result.error,
	}, nil
}

type queryValidationResult struct {
	valid         bool
	error         string
	operationType string
	operationName string
	depth         int
}

// forbiddenTypes are GraphQL types that contain sensitive data and should not be accessible via MCP queries.
// These types and their fields expose secret values that should not be returned to LLMs.
var forbiddenTypes = map[string]bool{
	"Secret":                           true, // The Secret type contains secret values
	"SecretValue":                      true, // SecretValue contains the actual secret data
	"SecretConnection":                 true, // Connection type that returns Secret nodes
	"SecretEdge":                       true, // Edge type that wraps Secret
	"DeploymentKey":                    true, // Contains the actual deployment key
	"CreateServiceAccountTokenPayload": true, // Contains the service account token secret
	"ServiceAccountToken":              true, // Service account token metadata (but secret field is blocked separately)
	"ServiceAccountTokenConnection":    true, // Connection type that returns ServiceAccountToken nodes
	"ServiceAccountTokenEdge":          true, // Edge type that wraps ServiceAccountToken
}

// checkForSecrets recursively checks if a selection set accesses any forbidden secret-related types.
// It validates against the GraphQL schema to ensure queries don't access Secret or SecretValue types.
func checkForSecrets(selectionSet ast.SelectionSet, schema *ast.Schema) (bool, string) {
	for _, selection := range selectionSet {
		switch sel := selection.(type) {
		case *ast.Field:
			// Check if this field's definition exists in the schema
			if sel.Definition != nil {
				// Check if the field returns a forbidden type
				typeName := getBaseTypeName(sel.Definition.Type)
				if forbiddenTypes[typeName] {
					return true, fmt.Sprintf("MCP security policy: field '%s' returns type '%s' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.", sel.Name, typeName)
				}
			}

			// Recursively check nested selections
			if len(sel.SelectionSet) > 0 {
				if found, reason := checkForSecrets(sel.SelectionSet, schema); found {
					return true, reason
				}
			}
		case *ast.InlineFragment:
			// Check if the inline fragment is on a forbidden type
			if sel.TypeCondition != "" && forbiddenTypes[sel.TypeCondition] {
				return true, fmt.Sprintf("MCP security policy: inline fragment on type '%s' which contains sensitive data that cannot be accessed via this interface", sel.TypeCondition)
			}
			// Check inline fragments recursively
			if found, reason := checkForSecrets(sel.SelectionSet, schema); found {
				return true, reason
			}
		case *ast.FragmentSpread:
			// Fragment spreads would need fragment definitions to be fully validated
			// For now, we flag any fragment that has "secret" in its name as a heuristic
			if strings.Contains(strings.ToLower(sel.Name), "secret") {
				return true, fmt.Sprintf("MCP security policy: fragment '%s' may access sensitive data that cannot be accessed via this interface", sel.Name)
			}
		}
	}
	return false, ""
}

// getBaseTypeName extracts the base type name from a GraphQL type, removing list and non-null wrappers.
func getBaseTypeName(t *ast.Type) string {
	if t.Elem != nil {
		return getBaseTypeName(t.Elem)
	}
	return t.Name()
}

// validateGraphQLQuery validates a GraphQL query against the schema.
func (t *toolContext) validateGraphQLQuery(reqCtx context.Context, query string) (*queryValidationResult, error) {
	// Fetch the cached and repaired schema
	schemaStr, err := t.getCachedSchema(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema: %w", err)
	}

	// Parse the schema (already repaired by getCachedSchema)
	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphql", Input: schemaStr})
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", gqlErr)
	}

	// Parse the query
	doc, errList := gqlparser.LoadQueryWithRules(schema, query, nil)
	if len(errList) > 0 {
		return &queryValidationResult{
			valid: false,
			error: errList.Error(),
		}, nil
	}

	// Check that we have at least one operation
	if len(doc.Operations) == 0 {
		return &queryValidationResult{
			valid: false,
			error: "no operations found in query",
		}, nil
	}

	// Check operation type - only allow queries
	// We only validate the first operation
	op := doc.Operations[0]
	if op.Operation != ast.Query {
		return &queryValidationResult{
			valid: false,
			error: fmt.Sprintf("only query operations are allowed, got: %s", op.Operation),
		}, nil
	}

	// Check query depth
	depth := calculateQueryDepth(op.SelectionSet, 0)
	if depth > maxQueryDepth {
		return &queryValidationResult{
			valid: false,
			error: fmt.Sprintf("query depth %d exceeds maximum allowed depth of %d", depth, maxQueryDepth),
		}, nil
	}

	// Check for forbidden secret-related types and fields
	if found, reason := checkForSecrets(op.SelectionSet, schema); found {
		return &queryValidationResult{
			valid: false,
			error: reason,
		}, nil
	}

	return &queryValidationResult{
		valid:         true,
		operationType: string(op.Operation),
		operationName: op.Name,
		depth:         depth,
	}, nil
}

func calculateQueryDepth(selectionSet ast.SelectionSet, currentDepth int) int {
	if len(selectionSet) == 0 {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, selection := range selectionSet {
		var childDepth int
		switch sel := selection.(type) {
		case *ast.Field:
			childDepth = calculateQueryDepth(sel.SelectionSet, currentDepth+1)
		case *ast.InlineFragment:
			childDepth = calculateQueryDepth(sel.SelectionSet, currentDepth)
		case *ast.FragmentSpread:
			// Fragment spreads would need to be resolved against fragment definitions
			// For simplicity, we count them as +1 depth
			childDepth = currentDepth + 1
		}
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth
}
