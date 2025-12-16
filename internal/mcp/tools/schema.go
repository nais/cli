// Package tools provides MCP tool implementations for Nais operations.
package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
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
		data, err := os.ReadFile(schemaFile)
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
		mcp.WithString("kind",
			mcp.Description("Filter by kind: 'OBJECT', 'INTERFACE', 'ENUM', 'UNION', 'INPUT_OBJECT', 'SCALAR', or 'all' (default: 'all')"),
		),
		mcp.WithString("search",
			mcp.Description("Filter type names containing this string (case-insensitive)"),
		),
	)
	s.AddTool(listTypesTool, ctx.handleSchemaListTypes)

	// Get type details tool
	getTypeTool := mcp.NewTool("schema_get_type",
		mcp.WithDescription("Get complete details about a GraphQL type: fields with their types, interfaces it implements, types that implement it (for interfaces), enum values, or union member types. Use this to understand the shape of data returned by queries."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The exact type name (e.g., 'Application', 'Team', 'DeploymentState')"),
		),
	)
	s.AddTool(getTypeTool, ctx.handleSchemaGetType)

	// List queries tool
	listQueriesTool := mcp.NewTool("schema_list_queries",
		mcp.WithDescription("List all available GraphQL query operations with their return types and number of arguments. These are the entry points for reading data from the Nais API."),
		mcp.WithString("search",
			mcp.Description("Filter query names or descriptions containing this string (case-insensitive)"),
		),
	)
	s.AddTool(listQueriesTool, ctx.handleSchemaListQueries)

	// List mutations tool
	listMutationsTool := mcp.NewTool("schema_list_mutations",
		mcp.WithDescription("List all available GraphQL mutation operations with their return types and number of arguments. Mutations are used to modify data (note: the MCP server currently only exposes read operations)."),
		mcp.WithString("search",
			mcp.Description("Filter mutation names or descriptions containing this string (case-insensitive)"),
		),
	)
	s.AddTool(listMutationsTool, ctx.handleSchemaListMutations)

	// Get field details tool
	getFieldTool := mcp.NewTool("schema_get_field",
		mcp.WithDescription("Get detailed information about a specific field including its arguments with types and defaults, return type, description, and deprecation status. Use 'Query' as the type to inspect query operations, or 'Mutation' for mutations."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("The type name containing the field (use 'Query' for root queries, 'Mutation' for root mutations, or any object type name)"),
		),
		mcp.WithString("field",
			mcp.Required(),
			mcp.Description("The field name to inspect"),
		),
	)
	s.AddTool(getFieldTool, ctx.handleSchemaGetField)

	// Get enum values tool
	getEnumTool := mcp.NewTool("schema_get_enum",
		mcp.WithDescription("Get all possible values for an enum type with their descriptions and deprecation status. Use this to understand valid values for enum fields (e.g., ApplicationState, DeploymentState)."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The enum type name (e.g., 'ApplicationState', 'TeamRole')"),
		),
	)
	s.AddTool(getEnumTool, ctx.handleSchemaGetEnum)

	// Search schema tool
	searchSchemaTool := mcp.NewTool("schema_search",
		mcp.WithDescription("Search across all schema types, fields, and enum values by name or description. Returns up to 50 matches. Use this to discover relevant types when you're not sure of exact names."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search term to match against names and descriptions (case-insensitive)"),
		),
	)
	s.AddTool(searchSchemaTool, ctx.handleSchemaSearch)

	// Get types implementing interface
	getImplementorsTool := mcp.NewTool("schema_get_implementors",
		mcp.WithDescription("Get all concrete types that implement a GraphQL interface. Use this to find all possible types when a query returns an interface type."),
		mcp.WithString("interface",
			mcp.Required(),
			mcp.Description("The interface name (e.g., 'Workload', 'Issue')"),
		),
	)
	s.AddTool(getImplementorsTool, ctx.handleSchemaGetImplementors)

	// Get union types
	getUnionTypesTool := mcp.NewTool("schema_get_union_types",
		mcp.WithDescription("Get all member types of a GraphQL union. Use this to understand what concrete types can be returned when a query returns a union type."),
		mcp.WithString("union",
			mcp.Required(),
			mcp.Description("The union type name"),
		),
	)
	s.AddTool(getUnionTypesTool, ctx.handleSchemaGetUnionTypes)
}

// handleSchemaListTypes handles the schema_list_types tool.
func (t *toolContext) handleSchemaListTypes(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.logger.Debug("Executing schema_list_types tool")

	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	t.logger.Debug("Fetching schema")
	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get schema", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}
	t.logger.Debug("Schema fetched", "size", len(schemaText))

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		t.logger.Error("Failed to parse schema", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}
	t.logger.Debug("Schema parsed successfully", "types", len(explorer.schema.Types))

	kind := strings.ToUpper(req.GetString("kind", "all"))
	search := strings.ToLower(req.GetString("search", ""))

	result := make(map[string][]string)

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

		kindKey := strings.ToLower(defKind) + "s"
		result[kindKey] = append(result[kindKey], name)
	}

	// Sort all lists
	for k := range result {
		sort.Strings(result[k])
	}

	return jsonResult(result)
}

// handleSchemaGetType handles the schema_get_type tool.
func (t *toolContext) handleSchemaGetType(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.logger.Debug("Executing schema_get_type tool")

	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	t.logger.Debug("Fetching schema for type lookup", "type_name", name)
	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		t.logger.Error("Failed to get schema", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		t.logger.Error("Failed to parse schema", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	def, ok := explorer.schema.Types[name]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("type %q not found", name)), nil
	}

	result := map[string]any{
		"name":        def.Name,
		"kind":        string(def.Kind),
		"description": def.Description,
	}

	// Add interfaces this type implements
	if len(def.Interfaces) > 0 {
		interfaces := append([]string(nil), def.Interfaces...)
		result["implements"] = interfaces
	}

	// Add fields for OBJECT, INTERFACE, INPUT_OBJECT
	if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
		result["fields"] = formatASTFields(def.Fields)
	}

	// Add enum values for ENUM
	if def.Kind == ast.Enum {
		var values []map[string]any
		for _, v := range def.EnumValues {
			value := map[string]any{
				"name": v.Name,
			}
			if v.Description != "" {
				value["description"] = v.Description
			}
			if v.Directives.ForName("deprecated") != nil {
				dep := v.Directives.ForName("deprecated")
				if reason := dep.Arguments.ForName("reason"); reason != nil {
					value["deprecated"] = reason.Value.Raw
				} else {
					value["deprecated"] = true
				}
			}
			values = append(values, value)
		}
		result["values"] = values
	}

	// Add types for UNION
	if def.Kind == ast.Union {
		types := append([]string(nil), def.Types...)
		result["types"] = types
	}

	// Add implementedBy for INTERFACE
	if def.Kind == ast.Interface {
		var implementedBy []string
		for typeName, typeDef := range explorer.schema.Types {
			for _, iface := range typeDef.Interfaces {
				if iface == name {
					implementedBy = append(implementedBy, typeName)
					break
				}
			}
		}
		sort.Strings(implementedBy)
		if len(implementedBy) > 0 {
			result["implementedBy"] = implementedBy
		}
	}

	return jsonResult(result)
}

// handleSchemaListQueries handles the schema_list_queries tool.
func (t *toolContext) handleSchemaListQueries(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	search := strings.ToLower(req.GetString("search", ""))

	queryType := explorer.schema.Query
	if queryType == nil {
		return mcp.NewToolResultError("Query type not found in schema"), nil
	}

	var queries []map[string]any
	for _, field := range queryType.Fields {
		if search == "" || strings.Contains(strings.ToLower(field.Name), search) || strings.Contains(strings.ToLower(field.Description), search) {
			queries = append(queries, map[string]any{
				"name":        field.Name,
				"returnType":  field.Type.String(),
				"description": truncate(field.Description, 150),
				"argCount":    len(field.Arguments),
			})
		}
	}

	// Sort by name
	sort.Slice(queries, func(i, j int) bool {
		return queries[i]["name"].(string) < queries[j]["name"].(string)
	})

	return jsonResult(queries)
}

// handleSchemaListMutations handles the schema_list_mutations tool.
func (t *toolContext) handleSchemaListMutations(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	search := strings.ToLower(req.GetString("search", ""))

	mutationType := explorer.schema.Mutation
	if mutationType == nil {
		return jsonResult([]map[string]any{})
	}

	var mutations []map[string]any
	for _, field := range mutationType.Fields {
		if search == "" || strings.Contains(strings.ToLower(field.Name), search) || strings.Contains(strings.ToLower(field.Description), search) {
			mutations = append(mutations, map[string]any{
				"name":        field.Name,
				"returnType":  field.Type.String(),
				"description": truncate(field.Description, 150),
				"argCount":    len(field.Arguments),
			})
		}
	}

	// Sort by name
	sort.Slice(mutations, func(i, j int) bool {
		return mutations[i]["name"].(string) < mutations[j]["name"].(string)
	})

	return jsonResult(mutations)
}

// handleSchemaGetField handles the schema_get_field tool.
func (t *toolContext) handleSchemaGetField(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	typeName, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	fieldName, err := req.RequireString("field")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	typeDef, ok := explorer.schema.Types[typeName]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("type %q not found", typeName)), nil
	}

	var field *ast.FieldDefinition
	for _, f := range typeDef.Fields {
		if f.Name == fieldName {
			field = f
			break
		}
	}

	if field == nil {
		return mcp.NewToolResultError(fmt.Sprintf("field %q not found on type %q", fieldName, typeName)), nil
	}

	result := map[string]any{
		"name":        field.Name,
		"type":        field.Type.String(),
		"description": field.Description,
	}

	// Check for deprecation
	if dep := field.Directives.ForName("deprecated"); dep != nil {
		if reason := dep.Arguments.ForName("reason"); reason != nil {
			result["deprecated"] = reason.Value.Raw
		} else {
			result["deprecated"] = true
		}
	}

	// Add arguments
	if len(field.Arguments) > 0 {
		var args []map[string]any
		for _, arg := range field.Arguments {
			argInfo := map[string]any{
				"name": arg.Name,
				"type": arg.Type.String(),
			}
			if arg.Description != "" {
				argInfo["description"] = arg.Description
			}
			if arg.DefaultValue != nil {
				argInfo["default"] = arg.DefaultValue.String()
			}
			args = append(args, argInfo)
		}
		result["args"] = args
	}

	return jsonResult(result)
}

// handleSchemaGetEnum handles the schema_get_enum tool.
func (t *toolContext) handleSchemaGetEnum(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	def, ok := explorer.schema.Types[name]
	if !ok || def.Kind != ast.Enum {
		return mcp.NewToolResultError(fmt.Sprintf("enum %q not found", name)), nil
	}

	var values []map[string]any
	for _, v := range def.EnumValues {
		value := map[string]any{
			"name": v.Name,
		}
		if v.Description != "" {
			value["description"] = v.Description
		}
		if dep := v.Directives.ForName("deprecated"); dep != nil {
			if reason := dep.Arguments.ForName("reason"); reason != nil {
				value["deprecated"] = reason.Value.Raw
			} else {
				value["deprecated"] = true
			}
		}
		values = append(values, value)
	}

	result := map[string]any{
		"name":        def.Name,
		"description": def.Description,
		"values":      values,
	}

	return jsonResult(result)
}

// handleSchemaSearch handles the schema_search tool.
func (t *toolContext) handleSchemaSearch(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	query = strings.ToLower(query)
	var results []map[string]any

	// Search types
	for name, def := range explorer.schema.Types {
		// Skip built-in types
		if strings.HasPrefix(name, "__") {
			continue
		}

		// Match type name or description
		if strings.Contains(strings.ToLower(name), query) || strings.Contains(strings.ToLower(def.Description), query) {
			results = append(results, map[string]any{
				"kind":        strings.ToLower(string(def.Kind)),
				"name":        name,
				"description": truncate(def.Description, 100),
			})
		}

		// Search fields
		for _, field := range def.Fields {
			if strings.Contains(strings.ToLower(field.Name), query) || strings.Contains(strings.ToLower(field.Description), query) {
				results = append(results, map[string]any{
					"kind":        "field",
					"type":        name,
					"name":        field.Name,
					"fieldType":   field.Type.String(),
					"description": truncate(field.Description, 100),
				})
			}
		}

		// Search enum values
		for _, v := range def.EnumValues {
			if strings.Contains(strings.ToLower(v.Name), query) || strings.Contains(strings.ToLower(v.Description), query) {
				results = append(results, map[string]any{
					"kind":        "enum_value",
					"enum":        name,
					"name":        v.Name,
					"description": truncate(v.Description, 100),
				})
			}
		}
	}

	// Sort results by kind then name
	sort.Slice(results, func(i, j int) bool {
		if results[i]["kind"].(string) != results[j]["kind"].(string) {
			return results[i]["kind"].(string) < results[j]["kind"].(string)
		}
		return results[i]["name"].(string) < results[j]["name"].(string)
	})

	// Limit results
	if len(results) > 50 {
		results = results[:50]
	}

	return jsonResult(map[string]any{
		"totalMatches": len(results),
		"results":      results,
	})
}

// handleSchemaGetImplementors handles the schema_get_implementors tool.
func (t *toolContext) handleSchemaGetImplementors(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	interfaceName, err := req.RequireString("interface")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	// Verify the interface exists
	def, ok := explorer.schema.Types[interfaceName]
	if !ok || def.Kind != ast.Interface {
		return mcp.NewToolResultError(fmt.Sprintf("interface %q not found", interfaceName)), nil
	}

	var implementors []map[string]any
	for typeName, typeDef := range explorer.schema.Types {
		for _, iface := range typeDef.Interfaces {
			if iface == interfaceName {
				implementors = append(implementors, map[string]any{
					"name":        typeName,
					"description": truncate(typeDef.Description, 100),
				})
				break
			}
		}
	}

	// Sort by name
	sort.Slice(implementors, func(i, j int) bool {
		return implementors[i]["name"].(string) < implementors[j]["name"].(string)
	})

	result := map[string]any{
		"interface":    interfaceName,
		"description":  def.Description,
		"implementors": implementors,
		"count":        len(implementors),
	}

	return jsonResult(result)
}

// handleSchemaGetUnionTypes handles the schema_get_union_types tool.
func (t *toolContext) handleSchemaGetUnionTypes(reqCtx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !t.rateLimiter.Allow() {
		return mcp.NewToolResultError("rate limit exceeded, please try again later"), nil
	}

	unionName, err := req.RequireString("union")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	schemaText, err := t.getSchema(reqCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	explorer, err := NewSchemaExplorer(schemaText)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse schema: %v", err)), nil
	}

	def, ok := explorer.schema.Types[unionName]
	if !ok || def.Kind != ast.Union {
		return mcp.NewToolResultError(fmt.Sprintf("union %q not found", unionName)), nil
	}

	var types []map[string]any
	for _, typeName := range def.Types {
		typeDef := explorer.schema.Types[typeName]
		typeInfo := map[string]any{
			"name": typeName,
		}
		if typeDef != nil {
			typeInfo["description"] = truncate(typeDef.Description, 100)
		}
		types = append(types, typeInfo)
	}

	result := map[string]any{
		"union":       unionName,
		"description": def.Description,
		"types":       types,
		"count":       len(types),
	}

	return jsonResult(result)
}

// formatASTFields formats ast.FieldDefinition list for output.
func formatASTFields(fields ast.FieldList) []map[string]any {
	var result []map[string]any
	for _, f := range fields {
		field := map[string]any{
			"name": f.Name,
			"type": f.Type.String(),
		}
		if f.Description != "" {
			field["description"] = truncate(f.Description, 150)
		}
		if dep := f.Directives.ForName("deprecated"); dep != nil {
			if reason := dep.Arguments.ForName("reason"); reason != nil {
				field["deprecated"] = reason.Value.Raw
			} else {
				field["deprecated"] = true
			}
		}
		if len(f.Arguments) > 0 {
			field["argCount"] = len(f.Arguments)
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

// jsonResult returns a JSON-formatted tool result.
func jsonResult(data any) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
