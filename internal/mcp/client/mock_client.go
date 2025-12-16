// Package client provides the GraphQL client interface for MCP operations.
package client

import (
	"context"

	gql "github.com/nais/cli/internal/naisapi/gql"
)

// Scenario represents a predefined mock scenario.
type Scenario string

const (
	// ScenarioHealthy represents an application without issues.
	ScenarioHealthy Scenario = "app-healthy"
	// ScenarioFailingInstance represents an application with failing instances.
	ScenarioFailingInstance Scenario = "app-failing-instance"
	// ScenarioCrashLoop represents an application in CrashLoopBackOff.
	ScenarioCrashLoop Scenario = "app-crashloop"
	// ScenarioPending represents an application with pending pods.
	ScenarioPending Scenario = "app-pending"
	// ScenarioHighCPU represents an application with high CPU utilization.
	ScenarioHighCPU Scenario = "app-high-cpu"
	// ScenarioVulnerable represents an application with critical vulnerabilities.
	ScenarioVulnerable Scenario = "app-vulnerable"
)

// MockClient implements the Client interface with deterministic mock data.
// This is always available (not behind a build tag) for use in tests.
type MockClient struct {
	scenario Scenario
}

// NewMockClient creates a new mock client with the specified scenario.
func NewMockClient(scenario Scenario) *MockClient {
	return &MockClient{scenario: scenario}
}

// SetScenario changes the current scenario.
func (c *MockClient) SetScenario(scenario Scenario) {
	c.scenario = scenario
}

// GetCurrentUser returns the current authenticated user (mock).
func (c *MockClient) GetCurrentUser(ctx context.Context) (*User, error) {
	return &User{
		Name:  "Mock User",
		Email: "mock-user@example.com",
	}, nil
}

// GetUserTeams returns the teams the current user is a member of (mock).
func (c *MockClient) GetUserTeams(ctx context.Context) ([]gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember, error) {
	return []gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMember{
		{
			Role: gql.TeamMemberRoleMember,
			Team: gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMemberTeam{
				Slug:    "team-alpha",
				Purpose: "Mock team for testing purposes",
			},
		},
		{
			Role: gql.TeamMemberRoleOwner,
			Team: gql.UserTeamsMeUserTeamsTeamMemberConnectionNodesTeamMemberTeam{
				Slug:    "team-beta",
				Purpose: "Another mock team for testing",
			},
		},
	}, nil
}

// GetSchema returns a mock GraphQL schema.
func (c *MockClient) GetSchema(ctx context.Context) (string, error) {
	return `# Mock GraphQL Schema
type Query {
  me: AuthenticatedUser
  teams(first: Int, filter: TeamFilter): TeamConnection!
  team(slug: String!): Team
  environments: EnvironmentConnection!
  search(filter: SearchFilter!, first: Int): SearchNodeConnection!
}

type AuthenticatedUser {
  email: String!
  teams: TeamMemberConnection!
  isAdmin: Boolean!
}

type Team {
  slug: String!
  purpose: String
  slackChannel: String
  members: TeamMemberConnection!
  applications(first: Int, filter: TeamApplicationsFilter): ApplicationConnection!
  workloads(first: Int, filter: TeamWorkloadsFilter): WorkloadConnection!
  jobs(first: Int, filter: TeamJobsFilter): JobConnection!
  deployments(first: Int): DeploymentConnection!
  issues(first: Int, filter: IssueFilter): IssueConnection!
  secrets(first: Int, filter: SecretFilter): SecretConnection!
  repositories(first: Int): RepositoryConnection!
  vulnerabilitySummary(filter: TeamVulnerabilitySummaryFilter): TeamVulnerabilitySummary!
  cost: TeamCost!
  environments: [TeamEnvironment!]!
  inventoryCounts: TeamInventoryCounts!
}

type Application {
  id: ID!
  name: String!
  state: ApplicationState!
  teamEnvironment: TeamEnvironment!
  image: ContainerImage!
  instances: ApplicationInstanceConnection!
  deployments(first: Int): DeploymentConnection!
  issues(first: Int, filter: IssueFilter): IssueConnection!
  ingresses: [Ingress!]!
  resources: ApplicationResources!
  networkPolicy: NetworkPolicy!
  authIntegrations: [ApplicationAuthIntegrations!]!
}

enum ApplicationState {
  RUNNING
  NOT_RUNNING
  UNKNOWN
}

enum ApplicationInstanceState {
  RUNNING
  FAILING
  UNKNOWN
}

type Job {
  id: ID!
  name: String!
  state: JobState!
  schedule: JobSchedule
  teamEnvironment: TeamEnvironment!
  image: ContainerImage!
  runs(first: Int): JobRunConnection!
  deployments(first: Int): DeploymentConnection!
  issues(first: Int, filter: IssueFilter): IssueConnection!
}

enum JobState {
  RUNNING
  COMPLETED
  FAILED
  UNKNOWN
}

type Deployment {
  id: ID!
  createdAt: Time!
  teamSlug: String!
  environmentName: String!
  repository: String
  deployerUsername: String
  commitSha: String
  triggerUrl: String
  resources: DeploymentResourceConnection!
  statuses: DeploymentStatusConnection!
}

type TeamVulnerabilitySummary {
  critical: Int!
  high: Int!
  medium: Int!
  low: Int!
  unassigned: Int!
  riskScore: Int!
  sbomCount: Int!
  coverage: Float!
}

type TeamCost {
  monthlySummary: TeamCostMonthlySummary!
}

type TeamCostMonthlySummary {
  sum: Float!
  series: [TeamCostMonthlySample!]!
}
`, nil
}

// GetConsoleURL returns a mock console URL.
func (c *MockClient) GetConsoleURL(ctx context.Context) (string, error) {
	return "https://console.nav.cloud.nais.io", nil
}
