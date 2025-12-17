package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func loadRealSchema(t *testing.T) *ast.Schema {
	t.Helper()

	// Load schema from the root of the project
	schemaPath := filepath.Join("..", "..", "..", "schema.graphql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	schemaStr := string(schemaBytes)

	// Remove builtin scalar redeclarations (same as the real implementation)
	filteredSchema := removeBuiltinScalars(schemaStr)

	// Parse the schema
	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphql", Input: filteredSchema})
	if gqlErr != nil {
		t.Fatalf("failed to parse schema: %v", gqlErr)
	}

	return schema
}

func TestCheckForSecrets_WithRealSchema(t *testing.T) {
	schema := loadRealSchema(t)

	tests := []struct {
		name          string
		query         string
		shouldBlock   bool
		expectedError string
	}{
		{
			name: "block query accessing Secret type via team.secrets",
			query: `
				query GetSecrets($slug: Slug!) {
					team(slug: $slug) {
						secrets(first: 10) {
							nodes {
								id
								name
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secrets' returns type 'SecretConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing Secret.values field",
			query: `
				query GetSecretValues($slug: Slug!) {
					team(slug: $slug) {
						secrets(first: 1) {
							nodes {
								name
								values {
									name
								}
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secrets' returns type 'SecretConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing SecretValue.value field",
			query: `
				query GetSecretValue($slug: Slug!) {
					team(slug: $slug) {
						secrets(first: 1) {
							nodes {
								values {
									name
									value
								}
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secrets' returns type 'SecretConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing secret via teamEnvironment.secret",
			query: `
				query GetSecret($slug: Slug!, $name: String!) {
					team(slug: $slug) {
						environments {
							environment {
								name
							}
							secret(name: $name) {
								id
								name
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secret' returns type 'Secret' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "allow query accessing activity log for secrets (activity log entries, not secret data)",
			query: `
				query GetSecretActivityLog($slug: Slug!) {
					team(slug: $slug) {
						activityLog(first: 10, filter: { activityTypes: [SECRET_CREATED, SECRET_DELETED] }) {
							nodes {
								__typename
								... on SecretCreatedActivityLogEntry {
									id
									actor
									createdAt
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "allow query accessing applications (no secrets)",
			query: `
				query GetApplications($slug: Slug!) {
					team(slug: $slug) {
						applications(first: 10) {
							nodes {
								name
								state
								teamEnvironment {
									environment {
										name
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "allow query accessing team info (no secrets)",
			query: `
				query GetTeam($slug: Slug!) {
					team(slug: $slug) {
						slug
						purpose
						slackChannel
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "allow query accessing user info (no secrets)",
			query: `
				query GetCurrentUser {
					me {
						... on User {
							name
							email
							teams(first: 50) {
								nodes {
									team {
										slug
									}
									role
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "allow query accessing application instances (no secrets)",
			query: `
				query GetAppInstances($slug: Slug!, $name: String!, $env: [String!]) {
					team(slug: $slug) {
						applications(filter: { name: $name, environments: $env }, first: 1) {
							nodes {
								name
								state
								instances {
									nodes {
										name
										restarts
										status {
											state
											message
										}
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "block query accessing application.secrets field",
			query: `
				query GetAppWithSecrets($slug: Slug!, $name: String!) {
					team(slug: $slug) {
						applications(filter: { name: $name }, first: 1) {
							nodes {
								name
								secrets(first: 10) {
									nodes {
										id
										name
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secrets' returns type 'SecretConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing job.secrets field",
			query: `
				query GetJobWithSecrets($slug: Slug!, $name: String!) {
					team(slug: $slug) {
						jobs(filter: { name: $name }, first: 1) {
							nodes {
								name
								secrets(first: 10) {
									nodes {
										id
										name
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'secrets' returns type 'SecretConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "allow query accessing deployments (no secrets)",
			query: `
				query GetDeployments($slug: Slug!) {
					team(slug: $slug) {
						deployments(first: 20) {
							nodes {
								createdAt
								repository
								commitSha
								statuses {
									nodes {
										state
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "allow nested query without secrets",
			query: `
				query GetComplexData($slug: Slug!) {
					team(slug: $slug) {
						applications(first: 10) {
							nodes {
								name
								image {
									name
									tag
									vulnerabilitySummary {
										critical
										high
										medium
										low
									}
								}
								teamEnvironment {
									environment {
										name
									}
								}
							}
						}
					}
				}
			`,
			shouldBlock: false,
		},
		{
			name: "block query accessing team.deploymentKey",
			query: `
				query GetDeploymentKey($slug: Slug!) {
					team(slug: $slug) {
						deploymentKey {
							id
							key
							created
							expires
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'deploymentKey' returns type 'DeploymentKey' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing DeploymentKey.key field",
			query: `
				query GetDeploymentKeyValue($slug: Slug!) {
					team(slug: $slug) {
						deploymentKey {
							key
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'deploymentKey' returns type 'DeploymentKey' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block mutation creating service account token (returns secret)",
			query: `
				mutation CreateToken($input: CreateServiceAccountTokenInput!) {
					createServiceAccountToken(input: $input) {
						serviceAccountToken {
							id
							name
						}
						secret
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'createServiceAccountToken' returns type 'CreateServiceAccountTokenPayload' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
		{
			name: "block query accessing service account tokens via Query.serviceAccounts",
			query: `
				query GetServiceAccountTokens {
					serviceAccounts(first: 10) {
						nodes {
							name
							tokens(first: 10) {
								nodes {
									id
									name
								}
							}
						}
					}
				}
			`,
			shouldBlock:   true,
			expectedError: "MCP security policy: field 'tokens' returns type 'ServiceAccountTokenConnection' which contains sensitive data that cannot be accessed via this interface. Use the Nais Console or CLI to manage secrets directly.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			doc, errList := gqlparser.LoadQueryWithRules(schema, tt.query, nil)
			if len(errList) > 0 {
				t.Fatalf("failed to parse query: %v", errList)
			}

			if len(doc.Operations) == 0 {
				t.Fatal("no operations found in query")
			}

			op := doc.Operations[0]

			// Check for secrets
			found, reason := checkForSecrets(op.SelectionSet, schema)

			if tt.shouldBlock && !found {
				t.Errorf("expected query to be blocked, but it was allowed")
			}

			if !tt.shouldBlock && found {
				t.Errorf("expected query to be allowed, but it was blocked with reason: %s", reason)
			}

			if tt.shouldBlock && found && tt.expectedError != "" {
				if reason != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, reason)
				}
			}
		})
	}
}

func TestGetBaseTypeName(t *testing.T) {
	tests := []struct {
		name     string
		typeDef  string
		expected string
	}{
		{
			name:     "simple type",
			typeDef:  "String",
			expected: "String",
		},
		{
			name:     "non-null type",
			typeDef:  "String!",
			expected: "String",
		},
		{
			name:     "list type",
			typeDef:  "[String]",
			expected: "String",
		},
		{
			name:     "non-null list",
			typeDef:  "[String]!",
			expected: "String",
		},
		{
			name:     "list of non-null",
			typeDef:  "[String!]",
			expected: "String",
		},
		{
			name:     "non-null list of non-null",
			typeDef:  "[String!]!",
			expected: "String",
		},
		{
			name:     "secret value list",
			typeDef:  "[SecretValue!]!",
			expected: "SecretValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse a minimal schema with the type
			schemaStr := `
				type Query {
					test: ` + tt.typeDef + `
				}
				type SecretValue {
					value: String!
				}
			`

			schema, err := gqlparser.LoadSchema(&ast.Source{
				Name:  "test.graphql",
				Input: schemaStr,
			})
			if err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			// Get the field definition
			queryType := schema.Types["Query"]
			if queryType == nil {
				t.Fatal("Query type not found")
			}

			testField := queryType.Fields.ForName("test")
			if testField == nil {
				t.Fatal("test field not found")
			}

			result := getBaseTypeName(testField.Type)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestForbiddenTypesAndFields(t *testing.T) {
	// Test that our forbidden types and fields are correct
	expectedTypes := []string{
		"Secret",
		"SecretValue",
		"SecretConnection",
		"SecretEdge",
		"DeploymentKey",
		"CreateServiceAccountTokenPayload",
		"ServiceAccountToken",
		"ServiceAccountTokenConnection",
		"ServiceAccountTokenEdge",
	}
	for _, typeName := range expectedTypes {
		if !forbiddenTypes[typeName] {
			t.Errorf("expected type %q to be forbidden", typeName)
		}
	}
}
