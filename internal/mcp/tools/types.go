package tools

// --- schema_list_types ---

type SchemaListTypesInput struct {
	Kind   string `json:"kind,omitempty" jsonschema_description:"Filter by kind: 'OBJECT', 'INTERFACE', 'ENUM', 'UNION', 'INPUT_OBJECT', 'SCALAR', or 'all' (default: 'all')"`
	Search string `json:"search,omitempty" jsonschema_description:"Filter type names containing this string (case-insensitive)"`
}

type SchemaListTypesOutput struct {
	Objects      []string `json:"objects,omitempty" jsonschema_description:"List of object type names"`
	Interfaces   []string `json:"interfaces,omitempty" jsonschema_description:"List of interface type names"`
	Enums        []string `json:"enums,omitempty" jsonschema_description:"List of enum type names"`
	Unions       []string `json:"unions,omitempty" jsonschema_description:"List of union type names"`
	InputObjects []string `json:"input_objects,omitempty" jsonschema_description:"List of input object type names"`
	Scalars      []string `json:"scalars,omitempty" jsonschema_description:"List of scalar type names"`
}

// --- schema_get_type ---

type SchemaGetTypeInput struct {
	Name string `json:"name" jsonschema:"required" jsonschema_description:"The exact type name (e.g., 'Application', 'Team', 'DeploymentState')"`
}

type SchemaFieldInfo struct {
	Name        string `json:"name" jsonschema_description:"Field name"`
	Type        string `json:"type" jsonschema_description:"Field type"`
	Description string `json:"description,omitempty" jsonschema_description:"Field description"`
	Deprecated  any    `json:"deprecated,omitempty" jsonschema_description:"Deprecation info if deprecated"`
}

type SchemaEnumValue struct {
	Name        string `json:"name" jsonschema_description:"Enum value name"`
	Description string `json:"description,omitempty" jsonschema_description:"Enum value description"`
	Deprecated  any    `json:"deprecated,omitempty" jsonschema_description:"Deprecation info if deprecated"`
}

type SchemaGetTypeOutput struct {
	Name          string            `json:"name" jsonschema_description:"Type name"`
	Kind          string            `json:"kind" jsonschema_description:"Type kind (OBJECT, INTERFACE, ENUM, etc.)"`
	Description   string            `json:"description,omitempty" jsonschema_description:"Type description"`
	Implements    []string          `json:"implements,omitempty" jsonschema_description:"Interfaces this type implements"`
	Fields        []SchemaFieldInfo `json:"fields,omitempty" jsonschema_description:"Fields on this type"`
	Values        []SchemaEnumValue `json:"values,omitempty" jsonschema_description:"Enum values (for ENUM types)"`
	Types         []string          `json:"types,omitempty" jsonschema_description:"Member types (for UNION types)"`
	ImplementedBy []string          `json:"implementedBy,omitempty" jsonschema_description:"Types implementing this interface"`
}

// --- schema_list_queries ---

type SchemaListQueriesInput struct {
	Search string `json:"search,omitempty" jsonschema_description:"Filter query names or descriptions containing this string (case-insensitive)"`
}

type SchemaOperationInfo struct {
	Name        string `json:"name" jsonschema_description:"Operation name"`
	ReturnType  string `json:"returnType" jsonschema_description:"Return type"`
	Description string `json:"description,omitempty" jsonschema_description:"Operation description"`
	ArgCount    int    `json:"argCount" jsonschema_description:"Number of arguments"`
}

// --- schema_list_mutations ---

type SchemaListMutationsInput struct {
	Search string `json:"search,omitempty" jsonschema_description:"Filter mutation names or descriptions containing this string (case-insensitive)"`
}

// --- schema_get_field ---

type SchemaGetFieldInput struct {
	Type  string `json:"type" jsonschema:"required" jsonschema_description:"The type name containing the field (use 'Query' for root queries, 'Mutation' for root mutations, or any object type name)"`
	Field string `json:"field" jsonschema:"required" jsonschema_description:"The field name to inspect"`
}

type SchemaArgumentInfo struct {
	Name        string `json:"name" jsonschema_description:"Argument name"`
	Type        string `json:"type" jsonschema_description:"Argument type"`
	Description string `json:"description,omitempty" jsonschema_description:"Argument description"`
	Default     string `json:"default,omitempty" jsonschema_description:"Default value if any"`
}

type SchemaGetFieldOutput struct {
	Name        string               `json:"name" jsonschema_description:"Field name"`
	Type        string               `json:"type" jsonschema_description:"Field return type"`
	Description string               `json:"description,omitempty" jsonschema_description:"Field description"`
	Deprecated  any                  `json:"deprecated,omitempty" jsonschema_description:"Deprecation info if deprecated"`
	Args        []SchemaArgumentInfo `json:"args,omitempty" jsonschema_description:"Field arguments"`
}

// --- schema_get_enum ---

type SchemaGetEnumInput struct {
	Name string `json:"name" jsonschema:"required" jsonschema_description:"The enum type name (e.g., 'ApplicationState', 'TeamRole')"`
}

type SchemaGetEnumOutput struct {
	Name        string            `json:"name" jsonschema_description:"Enum name"`
	Description string            `json:"description,omitempty" jsonschema_description:"Enum description"`
	Values      []SchemaEnumValue `json:"values" jsonschema_description:"Enum values"`
}

// --- schema_search ---

type SchemaSearchInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"Search term to match against names and descriptions (case-insensitive)"`
}

type SchemaSearchResult struct {
	Kind        string `json:"kind" jsonschema_description:"Result kind (object, interface, field, enum_value, etc.)"`
	Name        string `json:"name" jsonschema_description:"Name of the matched item"`
	Type        string `json:"type,omitempty" jsonschema_description:"Parent type (for fields)"`
	Enum        string `json:"enum,omitempty" jsonschema_description:"Parent enum (for enum values)"`
	FieldType   string `json:"fieldType,omitempty" jsonschema_description:"Field type (for fields)"`
	Description string `json:"description,omitempty" jsonschema_description:"Description"`
}

type SchemaSearchOutput struct {
	TotalMatches int                  `json:"totalMatches" jsonschema_description:"Total number of matches found"`
	Results      []SchemaSearchResult `json:"results" jsonschema_description:"Search results (max 50)"`
}

// --- schema_get_implementors ---

type SchemaGetImplementorsInput struct {
	Interface string `json:"interface" jsonschema:"required" jsonschema_description:"The interface name (e.g., 'Workload', 'Issue')"`
}

type SchemaImplementorInfo struct {
	Name        string `json:"name" jsonschema_description:"Type name"`
	Description string `json:"description,omitempty" jsonschema_description:"Type description"`
}

type SchemaGetImplementorsOutput struct {
	Interface    string                  `json:"interface" jsonschema_description:"Interface name"`
	Description  string                  `json:"description,omitempty" jsonschema_description:"Interface description"`
	Implementors []SchemaImplementorInfo `json:"implementors" jsonschema_description:"Types implementing this interface"`
	Count        int                     `json:"count" jsonschema_description:"Number of implementors"`
}

// --- schema_get_union_types ---

type SchemaGetUnionTypesInput struct {
	Union string `json:"union" jsonschema:"required" jsonschema_description:"The union type name"`
}

type SchemaUnionMember struct {
	Name        string `json:"name" jsonschema_description:"Member type name"`
	Description string `json:"description,omitempty" jsonschema_description:"Member type description"`
}

type SchemaGetUnionTypesOutput struct {
	Union       string              `json:"union" jsonschema_description:"Union name"`
	Description string              `json:"description,omitempty" jsonschema_description:"Union description"`
	Types       []SchemaUnionMember `json:"types" jsonschema_description:"Member types"`
	Count       int                 `json:"count" jsonschema_description:"Number of member types"`
}

// =============================================================================
// GraphQL Tool Types
// =============================================================================

// --- get_nais_context ---

type GetNaisContextInput struct{}

type NaisTeamInfo struct {
	Slug    string `json:"slug" jsonschema_description:"Team slug identifier"`
	Purpose string `json:"purpose,omitempty" jsonschema_description:"Team purpose"`
	Role    string `json:"role" jsonschema_description:"User's role in the team"`
}

type NaisUserInfo struct {
	Name string `json:"name" jsonschema_description:"User's name"`
}

type GetNaisContextOutput struct {
	User               NaisUserInfo      `json:"user" jsonschema_description:"Current user info"`
	Teams              []NaisTeamInfo    `json:"teams" jsonschema_description:"User's teams"`
	ConsoleBaseURL     string            `json:"console_base_url" jsonschema_description:"Base URL for Nais console"`
	ConsoleURLPatterns map[string]string `json:"console_url_patterns" jsonschema_description:"URL patterns for console pages"`
}

// --- execute_graphql ---

type ExecuteGraphQLInput struct {
	Query     string `json:"query" jsonschema:"required" jsonschema_description:"The GraphQL query to execute. Must be a query operation (not mutation or subscription)."`
	Variables string `json:"variables,omitempty" jsonschema_description:"JSON object containing variables for the query. Example: {\"slug\": \"my-team\", \"first\": 10}"`
}

// --- validate_graphql ---

type ValidateGraphQLInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"The GraphQL query to validate."`
}

type ValidateGraphQLOutput struct {
	Valid         bool   `json:"valid" jsonschema_description:"Whether the query is valid"`
	Error         string `json:"error,omitempty" jsonschema_description:"Validation error message if invalid"`
	OperationType string `json:"operationType,omitempty" jsonschema_description:"Type of operation (query, mutation, subscription)"`
	OperationName string `json:"operationName,omitempty" jsonschema_description:"Name of the operation if provided"`
	Depth         int    `json:"depth,omitempty" jsonschema_description:"Query depth"`
}
