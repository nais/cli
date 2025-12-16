// Package tools provides MCP tool implementations for Nais operations.
package tools

import (
	"strings"
	"testing"
)

const testSchema = `
"""
Directs the executor to include this field or fragment only when the if argument is true.
"""
directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

directive @deprecated(reason: String = "No longer supported") on FIELD_DEFINITION | ENUM_VALUE

"""
The application state.
"""
enum ApplicationState {
	"""
	The application is running.
	"""
	RUNNING
	"""
	The application is not running.
	"""
	NOT_RUNNING
	"""
	The application state is unknown.
	"""
	UNKNOWN
}

enum JobState {
	COMPLETED
	RUNNING
	FAILED
	UNKNOWN
}

"""
Types of scaling strategies.
"""
union ScalingStrategy = CPUScalingStrategy | KafkaLagScalingStrategy

"""
Authentication integrations for the application.
"""
union ApplicationAuthIntegrations = EntraIDAuthIntegration | IDPortenAuthIntegration

"""
Common interface for workloads.
"""
interface Workload {
	id: ID!
	name: String!
	teamEnvironment: TeamEnvironment!
}

"""
Common interface for persistence resources.
"""
interface Persistence {
	id: ID!
	name: String!
	team: Team!
}

"""
Input for filtering the applications of a team.
"""
input TeamApplicationsFilter {
	"""
	Filter by application name.
	"""
	name: String
	"""
	Filter by environments.
	"""
	environments: [String!]
}

input IssueFilter {
	resourceName: String
	resourceType: String
	environments: [String!]
	severity: String
}

type TeamEnvironment {
	name: String!
}

type Ingress {
	url: String!
}

type ApplicationInstanceConnection {
	nodes: [ApplicationInstance!]!
}

type ApplicationInstance {
	id: ID!
	name: String!
}

"""
A Nais application deployed to a team environment.
"""
type Application implements Workload {
	"""
	The globally unique ID of the application.
	"""
	id: ID!
	"""
	The name of the application.
	"""
	name: String!
	"""
	The team environment for the application.
	"""
	teamEnvironment: TeamEnvironment!
	"""
	The application state.
	"""
	state: ApplicationState!
	"""
	List of ingresses for the application.
	"""
	ingresses: [Ingress!]!
	"""
	The application instances.
	"""
	instances(
		"""
		Get the first n items in the connection.
		"""
		first: Int
		"""
		Get items after this cursor.
		"""
		after: String
	): ApplicationInstanceConnection!
	"""
	Old field that is deprecated.
	"""
	environment: TeamEnvironment! @deprecated(reason: "Use the teamEnvironment field instead.")
}

"""
A Nais team.
"""
type Team {
	"""
	The globally unique ID of the team.
	"""
	id: ID!
	"""
	Unique slug of the team.
	"""
	slug: String!
	"""
	Purpose of the team.
	"""
	purpose: String!
	"""
	Main Slack channel for the team.
	"""
	slackChannel: String!
	"""
	Nais applications owned by the team.
	"""
	applications(
		first: Int
		after: String
		orderBy: String
		filter: TeamApplicationsFilter
	): ApplicationConnection!
}

type ApplicationConnection {
	nodes: [Application!]!
}

type JobSchedule {
	expression: String!
}

type Job implements Workload {
	id: ID!
	name: String!
	teamEnvironment: TeamEnvironment!
	state: JobState!
	schedule: JobSchedule
}

type CPUScalingStrategy {
	threshold: Int!
}

type KafkaLagScalingStrategy {
	threshold: Int!
	consumerGroup: String!
	topicName: String!
}

type EntraIDAuthIntegration {
	name: String!
}

type IDPortenAuthIntegration {
	name: String!
}

type TeamConnection {
	nodes: [Team!]!
}

type EnvironmentConnection {
	nodes: [Environment!]!
}

type Environment {
	name: String!
}

type AuthenticatedUser {
	email: String!
}

type RestartApplicationPayload {
	success: Boolean!
}

input RestartApplicationInput {
	teamSlug: String!
	name: String!
	environment: String!
}

type AddTeamMemberPayload {
	success: Boolean!
}

input AddTeamMemberInput {
	teamSlug: String!
	email: String!
	role: String!
}

"""
The query root for the Nais GraphQL API.
"""
type Query {
	"""
	Get a list of teams.
	"""
	teams(
		first: Int
		after: String
		orderBy: String
		filter: String
	): TeamConnection!
	"""
	Get a team by its slug.
	"""
	team(
		slug: String!
	): Team!
	"""
	Get a list of environments.
	"""
	environments(
		orderBy: String
	): EnvironmentConnection!
	"""
	The currently authenticated user.
	"""
	me: AuthenticatedUser!
}

type Mutation {
	"""
	Restart an application.
	"""
	restartApplication(
		input: RestartApplicationInput!
	): RestartApplicationPayload!
	"""
	Add a team member.
	"""
	addTeamMember(
		input: AddTeamMemberInput!
	): AddTeamMemberPayload!
}
`

func TestSchemaExplorer_ParseEnums(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Check ApplicationState enum
	appState, ok := explorer.schema.Types["ApplicationState"]
	if !ok {
		t.Fatal("ApplicationState enum not found")
	}

	if appState.Description != "The application state." {
		t.Errorf("unexpected description: %q", appState.Description)
	}

	if len(appState.EnumValues) != 3 {
		t.Errorf("expected 3 values, got %d", len(appState.EnumValues))
	}

	// Check JobState enum (no description)
	jobState, ok := explorer.schema.Types["JobState"]
	if !ok {
		t.Fatal("JobState enum not found")
	}

	if len(jobState.EnumValues) != 4 {
		t.Errorf("expected 4 values, got %d", len(jobState.EnumValues))
	}
}

func TestSchemaExplorer_ParseUnions(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Check ScalingStrategy union
	scaling, ok := explorer.schema.Types["ScalingStrategy"]
	if !ok {
		t.Fatal("ScalingStrategy union not found")
	}

	if len(scaling.Types) != 2 {
		t.Errorf("expected 2 types, got %d", len(scaling.Types))
	}

	// Check ApplicationAuthIntegrations union
	auth, ok := explorer.schema.Types["ApplicationAuthIntegrations"]
	if !ok {
		t.Fatal("ApplicationAuthIntegrations union not found")
	}

	if len(auth.Types) != 2 {
		t.Errorf("expected 2 types, got %d", len(auth.Types))
	}
}

func TestSchemaExplorer_ParseInterfaces(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Check Workload interface
	workload, ok := explorer.schema.Types["Workload"]
	if !ok {
		t.Fatal("Workload interface not found")
	}

	if workload.Description != "Common interface for workloads." {
		t.Errorf("unexpected description: %q", workload.Description)
	}

	if len(workload.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(workload.Fields))
	}
}

func TestSchemaExplorer_ParseInputs(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Check TeamApplicationsFilter input
	filter, ok := explorer.schema.Types["TeamApplicationsFilter"]
	if !ok {
		t.Fatal("TeamApplicationsFilter input not found")
	}

	if filter.Description != "Input for filtering the applications of a team." {
		t.Errorf("unexpected description: %q", filter.Description)
	}

	if len(filter.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(filter.Fields))
	}

	// Check IssueFilter input (no description)
	issue, ok := explorer.schema.Types["IssueFilter"]
	if !ok {
		t.Fatal("IssueFilter input not found")
	}

	if len(issue.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(issue.Fields))
	}
}

func TestSchemaExplorer_ParseTypes(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Check Application type
	app, ok := explorer.schema.Types["Application"]
	if !ok {
		t.Fatal("Application type not found")
	}

	if app.Description != "A Nais application deployed to a team environment." {
		t.Errorf("unexpected description: %q", app.Description)
	}

	// Check implements
	if len(app.Interfaces) != 1 || app.Interfaces[0] != "Workload" {
		t.Errorf("expected [Workload], got %v", app.Interfaces)
	}

	// Check fields
	if len(app.Fields) < 5 {
		t.Errorf("expected at least 5 fields, got %d", len(app.Fields))
	}

	// Check Team type
	team, ok := explorer.schema.Types["Team"]
	if !ok {
		t.Fatal("Team type not found")
	}

	if len(team.Fields) < 4 {
		t.Errorf("expected at least 4 fields, got %d", len(team.Fields))
	}

	// Check Job type implements Workload
	job, ok := explorer.schema.Types["Job"]
	if !ok {
		t.Fatal("Job type not found")
	}

	if len(job.Interfaces) != 1 || job.Interfaces[0] != "Workload" {
		t.Errorf("expected [Workload], got %v", job.Interfaces)
	}
}

func TestSchemaExplorer_ParseQueries(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	queryType := explorer.schema.Query
	if queryType == nil {
		t.Fatal("Query type not found")
	}

	// Count only our defined queries (not built-in __schema, __type)
	var queryCount int
	for _, f := range queryType.Fields {
		if !strings.HasPrefix(f.Name, "__") {
			queryCount++
		}
	}
	if queryCount != 4 {
		t.Errorf("expected 4 queries, got %d", queryCount)
	}

	// Find teams query
	var teamsQuery *struct {
		Name        string
		Description string
		Args        int
	}
	for _, q := range queryType.Fields {
		if q.Name == "teams" {
			teamsQuery = &struct {
				Name        string
				Description string
				Args        int
			}{
				Name:        q.Name,
				Description: q.Description,
				Args:        len(q.Arguments),
			}
			break
		}
	}

	if teamsQuery == nil {
		t.Fatal("teams query not found")
	}

	if teamsQuery.Description != "Get a list of teams." {
		t.Errorf("unexpected description: %q", teamsQuery.Description)
	}

	if teamsQuery.Args != 4 {
		t.Errorf("expected 4 args, got %d", teamsQuery.Args)
	}

	// Find team query
	var teamQuery *struct {
		Name string
		Args int
	}
	for _, q := range queryType.Fields {
		if q.Name == "team" {
			teamQuery = &struct {
				Name string
				Args int
			}{
				Name: q.Name,
				Args: len(q.Arguments),
			}
			break
		}
	}

	if teamQuery == nil {
		t.Fatal("team query not found")
	}

	if teamQuery.Args != 1 {
		t.Errorf("expected 1 arg, got %d", teamQuery.Args)
	}
}

func TestSchemaExplorer_ParseMutations(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	mutationType := explorer.schema.Mutation
	if mutationType == nil {
		t.Fatal("Mutation type not found")
	}

	if len(mutationType.Fields) != 2 {
		t.Errorf("expected 2 mutations, got %d", len(mutationType.Fields))
	}

	// Find restartApplication mutation
	var restartMutation *struct {
		Name        string
		Description string
	}
	for _, m := range mutationType.Fields {
		if m.Name == "restartApplication" {
			restartMutation = &struct {
				Name        string
				Description string
			}{
				Name:        m.Name,
				Description: m.Description,
			}
			break
		}
	}

	if restartMutation == nil {
		t.Fatal("restartApplication mutation not found")
	}

	if restartMutation.Description != "Restart an application." {
		t.Errorf("unexpected description: %q", restartMutation.Description)
	}
}

func TestSchemaExplorer_ImplementedBy(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Count types implementing Workload
	implementers := make(map[string]bool)
	for typeName, typeDef := range explorer.schema.Types {
		for _, iface := range typeDef.Interfaces {
			if iface == "Workload" {
				implementers[typeName] = true
			}
		}
	}

	// Application and Job both implement Workload
	if !implementers["Application"] {
		t.Error("expected Application to implement Workload")
	}
	if !implementers["Job"] {
		t.Error("expected Job to implement Workload")
	}
	if len(implementers) != 2 {
		t.Errorf("expected 2 implementers, got %d", len(implementers))
	}
}

func TestSchemaExplorer_FieldWithArgs(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	app, ok := explorer.schema.Types["Application"]
	if !ok {
		t.Fatal("Application type not found")
	}

	// Find instances field which has args
	var instancesField *struct {
		Name string
		Args int
	}
	for _, f := range app.Fields {
		if f.Name == "instances" {
			instancesField = &struct {
				Name string
				Args int
			}{
				Name: f.Name,
				Args: len(f.Arguments),
			}
			break
		}
	}

	if instancesField == nil {
		t.Fatal("instances field not found")
	}

	if instancesField.Args != 2 {
		t.Errorf("expected 2 args, got %d", instancesField.Args)
	}
}

func TestSchemaExplorer_DeprecatedField(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	app, ok := explorer.schema.Types["Application"]
	if !ok {
		t.Fatal("Application type not found")
	}

	// Find deprecated environment field
	var envField *struct {
		Name       string
		Deprecated bool
		Reason     string
	}
	for _, f := range app.Fields {
		if f.Name == "environment" {
			dep := f.Directives.ForName("deprecated")
			if dep != nil {
				reason := ""
				if arg := dep.Arguments.ForName("reason"); arg != nil {
					reason = arg.Value.Raw
				}
				envField = &struct {
					Name       string
					Deprecated bool
					Reason     string
				}{
					Name:       f.Name,
					Deprecated: true,
					Reason:     reason,
				}
			}
			break
		}
	}

	if envField == nil {
		t.Fatal("environment field not found")
	}

	if !envField.Deprecated {
		t.Error("expected environment field to be deprecated")
	}

	if envField.Reason != "Use the teamEnvironment field instead." {
		t.Errorf("unexpected deprecation reason: %q", envField.Reason)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is a ..."},
		{"", 10, ""},
		{"multi\nline\ntext", 20, "multi line text"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestRemoveBuiltinScalars(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "removes Boolean scalar with description",
			input: `type Foo {
	bar: String!
}

"""
The ` + "`Boolean`" + ` scalar type represents ` + "`true`" + ` or ` + "`false`" + `.
"""
scalar Boolean

type Baz {
	qux: Boolean!
}
`,
			shouldContain:    []string{"type Foo", "type Baz", "qux: Boolean!"},
			shouldNotContain: []string{"scalar Boolean"},
		},
		{
			name: "removes multiple builtin scalars",
			input: `"""
The ` + "`Boolean`" + ` scalar type represents true or false.
"""
scalar Boolean

"""
The ` + "`Int`" + ` scalar type represents non-fractional signed whole numeric values.
"""
scalar Int

type MyType {
	flag: Boolean!
	count: Int!
}
`,
			shouldContain:    []string{"type MyType", "flag: Boolean!", "count: Int!"},
			shouldNotContain: []string{"scalar Boolean", "scalar Int"},
		},
		{
			name: "preserves non-builtin scalars",
			input: `"""
Custom scalar for dates.
"""
scalar Date

"""
The ` + "`Boolean`" + ` scalar type represents true or false.
"""
scalar Boolean

type Event {
	date: Date!
	active: Boolean!
}
`,
			shouldContain:    []string{"scalar Date", "type Event", "date: Date!"},
			shouldNotContain: []string{"scalar Boolean"},
		},
		{
			name: "handles ID scalar near IDPorten type",
			input: `"""
The ` + "`ID`" + ` scalar type represents a unique identifier.
"""
scalar ID

"""
ID-porten authentication.
"""
type IDPortenAuthIntegration {
	name: String!
}
`,
			shouldContain:    []string{"type IDPortenAuthIntegration", "ID-porten authentication"},
			shouldNotContain: []string{"scalar ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeBuiltinScalars(tt.input)
			for _, s := range tt.shouldContain {
				if !strings.Contains(result, s) {
					t.Errorf("removeBuiltinScalars() should contain %q, got:\n%s", s, result)
				}
			}
			for _, s := range tt.shouldNotContain {
				if strings.Contains(result, s) {
					t.Errorf("removeBuiltinScalars() should NOT contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}

func TestRemoveBuiltinScalars_ParsesSuccessfully(t *testing.T) {
	// Schema with builtin scalar redeclarations (like the real schema.graphql)
	schemaWithBuiltins := `
"""
The ` + "`Boolean`" + ` scalar type represents ` + "`true`" + ` or ` + "`false`" + `.
"""
scalar Boolean

"""
The ` + "`String`" + `scalar type represents textual data.
"""
scalar String

"""
The ` + "`Int`" + ` scalar type represents non-fractional signed whole numeric values.
"""
scalar Int

"""
The ` + "`Float`" + ` scalar type represents signed double-precision fractional values.
"""
scalar Float

"""
The ` + "`ID`" + ` scalar type represents a unique identifier.
"""
scalar ID

type Query {
	hello: String!
}

type User {
	id: ID!
	name: String!
	age: Int!
	score: Float!
	active: Boolean!
}
`

	// This should work without errors
	explorer, err := NewSchemaExplorer(schemaWithBuiltins)
	if err != nil {
		t.Fatalf("failed to parse schema with builtin scalars: %v", err)
	}

	// Verify the types are still accessible
	user, ok := explorer.schema.Types["User"]
	if !ok {
		t.Fatal("User type not found")
	}

	if len(user.Fields) != 5 {
		t.Errorf("expected 5 fields on User, got %d", len(user.Fields))
	}
}

func TestFormatASTFields(t *testing.T) {
	explorer, err := NewSchemaExplorer(testSchema)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	app := explorer.schema.Types["Application"]
	result := formatASTFields(app.Fields)

	if len(result) == 0 {
		t.Error("expected formatted fields")
	}

	// Check that fields have expected keys
	for _, f := range result {
		if _, ok := f["name"]; !ok {
			t.Error("expected name key in field")
		}
		if _, ok := f["type"]; !ok {
			t.Error("expected type key in field")
		}
	}

	// Check deprecated field
	var foundDeprecated bool
	for _, f := range result {
		if f["name"] == "environment" {
			if _, ok := f["deprecated"]; ok {
				foundDeprecated = true
			}
		}
	}
	if !foundDeprecated {
		t.Error("expected to find deprecated field with deprecated key")
	}
}
