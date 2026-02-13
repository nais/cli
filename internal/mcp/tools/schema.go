// Package tools provides MCP tool implementations for Nais operations.
package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// SchemaExplorer provides efficient exploration of the GraphQL schema using gqlparser.
type SchemaExplorer struct {
	schema *ast.Schema
}

// NewSchemaExplorer creates a new schema explorer from raw schema text.
func NewSchemaExplorer(schemaText string) (*SchemaExplorer, error) {
	// Remove built-in scalar redeclarations that conflict with gqlparser's built-ins
	filteredSchema := removeBuiltinScalars(schemaText)

	source := &ast.Source{
		Name:  "schema.graphql",
		Input: filteredSchema,
	}

	schema, err := gqlparser.LoadSchema(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &SchemaExplorer{schema: schema}, nil
}

// removeBuiltinScalars removes scalar definitions for built-in GraphQL types
// (Boolean, String, Int, Float, ID) that gqlparser already defines internally.
// These redeclarations in the schema cause "Cannot redeclare type" errors.
func removeBuiltinScalars(schema string) string {
	builtins := map[string]bool{
		"Boolean": true,
		"String":  true,
		"Int":     true,
		"Float":   true,
		"ID":      true,
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(schema))

	var descriptionLines []string
	inDescription := false

	for scanner.Scan() {
		line := scanner.Text()

		// Track if we're entering a description block
		if strings.HasPrefix(strings.TrimSpace(line), `"""`) {
			if !inDescription {
				// Starting a description block
				inDescription = true
				descriptionLines = []string{line}

				// Check if it ends on the same line (single-line description)
				trimmed := strings.TrimSpace(line)
				if len(trimmed) > 6 && strings.HasSuffix(trimmed, `"""`) {
					inDescription = false
				}
				continue
			} else {
				// Ending a description block
				inDescription = false
				descriptionLines = append(descriptionLines, line)
				continue
			}
		}

		if inDescription {
			descriptionLines = append(descriptionLines, line)
			continue
		}

		// Check if this is a scalar line for a builtin type
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "scalar ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				scalarName := parts[1]
				if builtins[scalarName] {
					// Skip this scalar and the description we just collected
					descriptionLines = nil
					continue
				}
			}
		}

		// If we have pending description lines, write them now
		if len(descriptionLines) > 0 {
			for _, descLine := range descriptionLines {
				result.WriteString(descLine)
				result.WriteString("\n")
			}
			descriptionLines = nil
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// getSchema retrieves the schema either from a local file (if NAIS_SCHEMA_FILE env var is set)
// or via the toolContext's cached schema (which is fetched once and repaired).
func (t *toolContext) getSchema(reqCtx context.Context) (string, error) {
	// Check for local schema file override
	if schemaFile := os.Getenv("NAIS_SCHEMA_FILE"); schemaFile != "" {
		data, err := os.ReadFile(filepath.Clean(schemaFile))
		if err != nil {
			return "", fmt.Errorf("failed to read schema file %s: %w", schemaFile, err)
		}
		// Still need to repair the local schema
		return removeBuiltinScalars(string(data)), nil
	}

	// Use cached schema (fetched once, repaired, and stored)
	return t.getCachedSchema(reqCtx)
}

// registerSchemaTools registers schema exploration tools.
func registerSchemaTools(s *server.MCPServer, ctx *toolContext) {
	// List all types tool
	listTypesTool := mcp.NewTool("schema_list_types",
		mcp.WithDescription("List all types in the Nais GraphQL API schema, grouped by kind. Use this to explore available data types before querying specific type details. Useful for understanding the API structure."),
		mcp.WithInputSchema[SchemaListTypesInput](),
		mcp.WithOutputSchema[SchemaListTypesOutput](),
	)
	s.AddTool(listTypesTool, mcp.NewStructuredToolHandler(ctx.handleSchemaListTypes))

	// Get type details tool
	getTypeTool := mcp.NewTool("schema_get_type",
		mcp.WithDescription("Get complete details about a GraphQL type: fields with their types, interfaces it implements, types that implement it (for interfaces), enum values, or union member types. Use this to understand the shape of data returned by queries."),
		mcp.WithInputSchema[SchemaGetTypeInput](),
		mcp.WithOutputSchema[SchemaGetTypeOutput](),
	)
	s.AddTool(getTypeTool, mcp.NewStructuredToolHandler(ctx.handleSchemaGetType))

	// List queries tool
	listQueriesTool := mcp.NewTool("schema_list_queries",
		mcp.WithDescription("List all available GraphQL query operations with their return types and number of arguments. These are the entry points for reading data from the Nais API."),
		mcp.WithInputSchema[SchemaListQueriesInput](),
		mcp.WithOutputSchema[[]SchemaOperationInfo](),
	)
	s.AddTool(listQueriesTool, mcp.NewStructuredToolHandler(ctx.handleSchemaListQueries))

	// List mutations tool
	listMutationsTool := mcp.NewTool("schema_list_mutations",
		mcp.WithDescription("List all available GraphQL mutation operations with their return types and number of arguments. Mutations are used to modify data (note: the MCP server currently only exposes read operations)."),
		mcp.WithInputSchema[SchemaListMutationsInput](),
		mcp.WithOutputSchema[[]SchemaOperationInfo](),
	)
	s.AddTool(listMutationsTool, mcp.NewStructuredToolHandler(ctx.handleSchemaListMutations))

	// Get field details tool
	getFieldTool := mcp.NewTool("schema_get_field",
		mcp.WithDescription("Get detailed information about a specific field including its arguments with types and defaults, return type, description, and deprecation status. Use 'Query' as the type to inspect query operations, or 'Mutation' for mutations."),
		mcp.WithInputSchema[SchemaGetFieldInput](),
		mcp.WithOutputSchema[SchemaGetFieldOutput](),
	)
	s.AddTool(getFieldTool, mcp.NewStructuredToolHandler(ctx.handleSchemaGetField))

	// Get enum values tool
	getEnumTool := mcp.NewTool("schema_get_enum",
		mcp.WithDescription("Get all possible values for an enum type with their descriptions and deprecation status. Use this to understand valid values for enum fields (e.g., ApplicationState, DeploymentState)."),
		mcp.WithInputSchema[SchemaGetEnumInput](),
		mcp.WithOutputSchema[SchemaGetEnumOutput](),
	)
	s.AddTool(getEnumTool, mcp.NewStructuredToolHandler(ctx.handleSchemaGetEnum))

	// Search schema tool
	searchSchemaTool := mcp.NewTool("schema_search",
		mcp.WithDescription("Search across all schema types, fields, and enum values by name or description. Returns up to 50 matches. Use this to discover relevant types when you're not sure of exact names."),
		mcp.WithInputSchema[SchemaSearchInput](),
		mcp.WithOutputSchema[SchemaSearchOutput](),
	)
	s.AddTool(searchSchemaTool, mcp.NewStructuredToolHandler(ctx.handleSchemaSearch))

	// Get types implementing interface
	getImplementorsTool := mcp.NewTool("schema_get_implementors",
		mcp.WithDescription("Get all concrete types that implement a GraphQL interface. Use this to find all possible types when a query returns an interface type."),
		mcp.WithInputSchema[SchemaGetImplementorsInput](),
		mcp.WithOutputSchema[SchemaGetImplementorsOutput](),
	)
	s.AddTool(getImplementorsTool, mcp.NewStructuredToolHandler(ctx.handleSchemaGetImplementors))

	// Get union types
	getUnionTypesTool := mcp.NewTool("schema_get_union_types",
		mcp.WithDescription("Get all member types of a GraphQL union. Use this to understand what concrete types can be returned when a query returns a union type."),
		mcp.WithInputSchema[SchemaGetUnionTypesInput](),
		mcp.WithOutputSchema[SchemaGetUnionTypesOutput](),
	)
	s.AddTool(getUnionTypesTool, mcp.NewStructuredToolHandler(ctx.handleSchemaGetUnionTypes))
}

// handleSchemaListTypes handles the schema_list_types tool.
func (t *toolContext) handleSchemaListTypes(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaListTypesInput,
) (SchemaListTypesOutput, error) {
	t.logger.Debug("Executing schema_list_types tool")

	if !t.rateLimiter.Allow() {
		return SchemaListTypesOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	t.logger.Debug("Fetching schema")
	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get schema", "error", err)
		return SchemaListTypesOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}
	t.logger.Debug("Schema fetched", "size", len(schemaText))

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		t.logger.Error("Failed to parse schema", "error", err)
		return SchemaListTypesOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}
	t.logger.Debug("Schema parsed successfully", "types", len(explorer.schema.Types))

	kind := strings.ToUpper(args.Kind)
	if kind == "" {
		kind = "ALL"
	}
	search := strings.ToLower(args.Search)

	var output SchemaListTypesOutput

	for name, def := range explorer.schema.Types {
		// Skip built-in types
		if strings.HasPrefix(name, "__") {
			continue
		}

		// Filter by kind
		defKind := string(def.Kind)
		if kind != "ALL" && defKind != kind {
			continue
		}

		// Filter by search
		if search != "" && !strings.Contains(strings.ToLower(name), search) {
			continue
		}

		switch def.Kind {
		case ast.Object:
			output.Objects = append(output.Objects, name)
		case ast.Interface:
			output.Interfaces = append(output.Interfaces, name)
		case ast.Enum:
			output.Enums = append(output.Enums, name)
		case ast.Union:
			output.Unions = append(output.Unions, name)
		case ast.InputObject:
			output.InputObjects = append(output.InputObjects, name)
		case ast.Scalar:
			output.Scalars = append(output.Scalars, name)
		}
	}

	// Sort all lists
	sort.Strings(output.Objects)
	sort.Strings(output.Interfaces)
	sort.Strings(output.Enums)
	sort.Strings(output.Unions)
	sort.Strings(output.InputObjects)
	sort.Strings(output.Scalars)

	return output, nil
}

// handleSchemaGetType handles the schema_get_type tool.
func (t *toolContext) handleSchemaGetType(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaGetTypeInput,
) (SchemaGetTypeOutput, error) {
	t.logger.Debug("Executing schema_get_type tool")

	if !t.rateLimiter.Allow() {
		return SchemaGetTypeOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	t.logger.Debug("Fetching schema for type lookup", "type_name", args.Name)
	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get schema", "error", err)
		return SchemaGetTypeOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		t.logger.Error("Failed to parse schema", "error", err)
		return SchemaGetTypeOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	def, ok := explorer.schema.Types[args.Name]
	if !ok {
		return SchemaGetTypeOutput{}, fmt.Errorf("type %q not found", args.Name)
	}

	output := SchemaGetTypeOutput{
		Name:        def.Name,
		Kind:        string(def.Kind),
		Description: def.Description,
	}

	// Add interfaces this type implements
	if len(def.Interfaces) > 0 {
		output.Implements = append([]string(nil), def.Interfaces...)
	}

	// Add fields for OBJECT, INTERFACE, INPUT_OBJECT
	if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
		output.Fields = formatASTFieldsTyped(def.Fields)
	}

	// Add enum values for ENUM
	if def.Kind == ast.Enum {
		for _, v := range def.EnumValues {
			value := SchemaEnumValue{
				Name:        v.Name,
				Description: v.Description,
			}
			if dep := v.Directives.ForName("deprecated"); dep != nil {
				if reason := dep.Arguments.ForName("reason"); reason != nil {
					value.Deprecated = reason.Value.Raw
				} else {
					value.Deprecated = true
				}
			}
			output.Values = append(output.Values, value)
		}
	}

	// Add types for UNION
	if def.Kind == ast.Union {
		output.Types = append([]string(nil), def.Types...)
	}

	// Add implementedBy for INTERFACE
	if def.Kind == ast.Interface {
		var implementedBy []string
		for typeName, typeDef := range explorer.schema.Types {
			for _, iface := range typeDef.Interfaces {
				if iface == args.Name {
					implementedBy = append(implementedBy, typeName)
					break
				}
			}
		}
		sort.Strings(implementedBy)
		if len(implementedBy) > 0 {
			output.ImplementedBy = implementedBy
		}
	}

	return output, nil
}

// handleSchemaListQueries handles the schema_list_queries tool.
func (t *toolContext) handleSchemaListQueries(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaListQueriesInput,
) ([]SchemaOperationInfo, error) {
	if !t.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	search := strings.ToLower(args.Search)

	queryType := explorer.schema.Query
	if queryType == nil {
		return nil, fmt.Errorf("query type not found in schema")
	}

	var queries []SchemaOperationInfo
	for _, field := range queryType.Fields {
		if search == "" || strings.Contains(strings.ToLower(field.Name), search) || strings.Contains(strings.ToLower(field.Description), search) {
			queries = append(queries, SchemaOperationInfo{
				Name:        field.Name,
				ReturnType:  field.Type.String(),
				Description: truncate(field.Description, 150),
				ArgCount:    len(field.Arguments),
			})
		}
	}

	// Sort by name
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Name < queries[j].Name
	})

	return queries, nil
}

// handleSchemaListMutations handles the schema_list_mutations tool.
func (t *toolContext) handleSchemaListMutations(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaListMutationsInput,
) ([]SchemaOperationInfo, error) {
	if !t.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	search := strings.ToLower(args.Search)

	mutationType := explorer.schema.Mutation
	if mutationType == nil {
		return []SchemaOperationInfo{}, nil
	}

	var mutations []SchemaOperationInfo
	for _, field := range mutationType.Fields {
		if search == "" || strings.Contains(strings.ToLower(field.Name), search) || strings.Contains(strings.ToLower(field.Description), search) {
			mutations = append(mutations, SchemaOperationInfo{
				Name:        field.Name,
				ReturnType:  field.Type.String(),
				Description: truncate(field.Description, 150),
				ArgCount:    len(field.Arguments),
			})
		}
	}

	// Sort by name
	sort.Slice(mutations, func(i, j int) bool {
		return mutations[i].Name < mutations[j].Name
	})

	return mutations, nil
}

// handleSchemaGetField handles the schema_get_field tool.
func (t *toolContext) handleSchemaGetField(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaGetFieldInput,
) (SchemaGetFieldOutput, error) {
	if !t.rateLimiter.Allow() {
		return SchemaGetFieldOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return SchemaGetFieldOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return SchemaGetFieldOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	typeDef, ok := explorer.schema.Types[args.Type]
	if !ok {
		return SchemaGetFieldOutput{}, fmt.Errorf("type %q not found", args.Type)
	}

	var field *ast.FieldDefinition
	for _, f := range typeDef.Fields {
		if f.Name == args.Field {
			field = f
			break
		}
	}

	if field == nil {
		return SchemaGetFieldOutput{}, fmt.Errorf("field %q not found on type %q", args.Field, args.Type)
	}

	output := SchemaGetFieldOutput{
		Name:        field.Name,
		Type:        field.Type.String(),
		Description: field.Description,
	}

	// Check for deprecation
	if dep := field.Directives.ForName("deprecated"); dep != nil {
		if reason := dep.Arguments.ForName("reason"); reason != nil {
			output.Deprecated = reason.Value.Raw
		} else {
			output.Deprecated = true
		}
	}

	// Add arguments
	if len(field.Arguments) > 0 {
		for _, arg := range field.Arguments {
			argInfo := SchemaArgumentInfo{
				Name:        arg.Name,
				Type:        arg.Type.String(),
				Description: arg.Description,
			}
			if arg.DefaultValue != nil {
				argInfo.Default = arg.DefaultValue.String()
			}
			output.Args = append(output.Args, argInfo)
		}
	}

	return output, nil
}

// handleSchemaGetEnum handles the schema_get_enum tool.
func (t *toolContext) handleSchemaGetEnum(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaGetEnumInput,
) (SchemaGetEnumOutput, error) {
	if !t.rateLimiter.Allow() {
		return SchemaGetEnumOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return SchemaGetEnumOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return SchemaGetEnumOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	def, ok := explorer.schema.Types[args.Name]
	if !ok || def.Kind != ast.Enum {
		return SchemaGetEnumOutput{}, fmt.Errorf("enum %q not found", args.Name)
	}

	output := SchemaGetEnumOutput{
		Name:        def.Name,
		Description: def.Description,
	}

	for _, v := range def.EnumValues {
		value := SchemaEnumValue{
			Name:        v.Name,
			Description: v.Description,
		}
		if dep := v.Directives.ForName("deprecated"); dep != nil {
			if reason := dep.Arguments.ForName("reason"); reason != nil {
				value.Deprecated = reason.Value.Raw
			} else {
				value.Deprecated = true
			}
		}
		output.Values = append(output.Values, value)
	}

	return output, nil
}

// handleSchemaSearch handles the schema_search tool.
func (t *toolContext) handleSchemaSearch(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaSearchInput,
) (SchemaSearchOutput, error) {
	if !t.rateLimiter.Allow() {
		return SchemaSearchOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return SchemaSearchOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return SchemaSearchOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	query := strings.ToLower(args.Query)
	var results []SchemaSearchResult

	// Search types
	for name, def := range explorer.schema.Types {
		// Skip built-in types
		if strings.HasPrefix(name, "__") {
			continue
		}

		// Match type name or description
		if strings.Contains(strings.ToLower(name), query) || strings.Contains(strings.ToLower(def.Description), query) {
			results = append(results, SchemaSearchResult{
				Kind:        strings.ToLower(string(def.Kind)),
				Name:        name,
				Description: truncate(def.Description, 100),
			})
		}

		// Search fields
		for _, field := range def.Fields {
			if strings.Contains(strings.ToLower(field.Name), query) || strings.Contains(strings.ToLower(field.Description), query) {
				results = append(results, SchemaSearchResult{
					Kind:        "field",
					Type:        name,
					Name:        field.Name,
					FieldType:   field.Type.String(),
					Description: truncate(field.Description, 100),
				})
			}
		}

		// Search enum values
		for _, v := range def.EnumValues {
			if strings.Contains(strings.ToLower(v.Name), query) || strings.Contains(strings.ToLower(v.Description), query) {
				results = append(results, SchemaSearchResult{
					Kind:        "enum_value",
					Enum:        name,
					Name:        v.Name,
					Description: truncate(v.Description, 100),
				})
			}
		}
	}

	// Sort results by kind then name
	sort.Slice(results, func(i, j int) bool {
		if results[i].Kind != results[j].Kind {
			return results[i].Kind < results[j].Kind
		}
		return results[i].Name < results[j].Name
	})

	// Limit results
	if len(results) > 50 {
		results = results[:50]
	}

	return SchemaSearchOutput{
		TotalMatches: len(results),
		Results:      results,
	}, nil
}

// handleSchemaGetImplementors handles the schema_get_implementors tool.
func (t *toolContext) handleSchemaGetImplementors(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaGetImplementorsInput,
) (SchemaGetImplementorsOutput, error) {
	if !t.rateLimiter.Allow() {
		return SchemaGetImplementorsOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return SchemaGetImplementorsOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return SchemaGetImplementorsOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Verify the interface exists
	def, ok := explorer.schema.Types[args.Interface]
	if !ok || def.Kind != ast.Interface {
		return SchemaGetImplementorsOutput{}, fmt.Errorf("interface %q not found", args.Interface)
	}

	var implementors []SchemaImplementorInfo
	for typeName, typeDef := range explorer.schema.Types {
		for _, iface := range typeDef.Interfaces {
			if iface == args.Interface {
				implementors = append(implementors, SchemaImplementorInfo{
					Name:        typeName,
					Description: truncate(typeDef.Description, 100),
				})
				break
			}
		}
	}

	// Sort by name
	sort.Slice(implementors, func(i, j int) bool {
		return implementors[i].Name < implementors[j].Name
	})

	return SchemaGetImplementorsOutput{
		Interface:    args.Interface,
		Description:  def.Description,
		Implementors: implementors,
		Count:        len(implementors),
	}, nil
}

// handleSchemaGetUnionTypes handles the schema_get_union_types tool.
func (t *toolContext) handleSchemaGetUnionTypes(
	reqCtx context.Context,
	req mcp.CallToolRequest,
	args SchemaGetUnionTypesInput,
) (SchemaGetUnionTypesOutput, error) {
	if !t.rateLimiter.Allow() {
		return SchemaGetUnionTypesOutput{}, fmt.Errorf("rate limit exceeded, please try again later")
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return SchemaGetUnionTypesOutput{}, fmt.Errorf("failed to get schema: %w", err)
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return SchemaGetUnionTypesOutput{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	def, ok := explorer.schema.Types[args.Union]
	if !ok || def.Kind != ast.Union {
		return SchemaGetUnionTypesOutput{}, fmt.Errorf("union %q not found", args.Union)
	}

	var types []SchemaUnionMember
	for _, typeName := range def.Types {
		typeDef := explorer.schema.Types[typeName]
		member := SchemaUnionMember{
			Name: typeName,
		}
		if typeDef != nil {
			member.Description = truncate(typeDef.Description, 100)
		}
		types = append(types, member)
	}

	return SchemaGetUnionTypesOutput{
		Union:       args.Union,
		Description: def.Description,
		Types:       types,
		Count:       len(types),
	}, nil
}

// formatASTFieldsTyped converts AST fields to typed SchemaFieldInfo slice.
func formatASTFieldsTyped(fields ast.FieldList) []SchemaFieldInfo {
	var result []SchemaFieldInfo
	for _, f := range fields {
		field := SchemaFieldInfo{
			Name:        f.Name,
			Type:        f.Type.String(),
			Description: truncate(f.Description, 150),
		}
		if dep := f.Directives.ForName("deprecated"); dep != nil {
			if reason := dep.Arguments.ForName("reason"); reason != nil {
				field.Deprecated = reason.Value.Raw
			} else {
				field.Deprecated = true
			}
		}
		result = append(result, field)
	}
	return result
}

// truncate truncates a string to the specified length.
func truncate(s string, length int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
