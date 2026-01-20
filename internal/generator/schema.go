package generator

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FormatSchema converts an OpenAPI schema into markdown format.
// indent controls the indentation level (each level = 2 spaces).
// maxDepth limits recursion depth to prevent stack overflow on circular references.
// Returns a markdown-formatted string representation of the schema.
func FormatSchema(schema *openapi3.Schema, indent, maxDepth int) string {
	if schema == nil {
		return ""
	}

	if maxDepth <= 0 {
		prefix := strings.Repeat("  ", indent)
		return fmt.Sprintf("%s- *(max depth reached)*\n", prefix)
	}

	var result strings.Builder
	prefix := strings.Repeat("  ", indent)

	// Handle schema composition (oneOf, anyOf, allOf)
	if len(schema.OneOf) > 0 {
		formatSchemaComposition(&result, "oneOf", "one of the following", schema.OneOf, prefix, indent, maxDepth)
		return result.String()
	}

	if len(schema.AnyOf) > 0 {
		formatSchemaComposition(&result, "anyOf", "any of the following", schema.AnyOf, prefix, indent, maxDepth)
		return result.String()
	}

	if len(schema.AllOf) > 0 {
		formatSchemaComposition(&result, "allOf", "all of the following", schema.AllOf, prefix, indent, maxDepth)
		return result.String()
	}

	// Handle object type
	if schema.Type.Is("object") {
		formatObjectSchema(&result, schema, prefix, indent, maxDepth)
		return result.String()
	}

	// Handle array type
	if schema.Type.Is("array") {
		formatArraySchema(&result, schema, prefix, indent, maxDepth)
		return result.String()
	}

	// Handle primitive types
	if schema.Type.Slice() != nil {
		formatPrimitiveSchema(&result, schema, prefix)
		return result.String()
	}

	return result.String()
}

// formatSchemaComposition formats oneOf/anyOf/allOf schemas.
func formatSchemaComposition(result *strings.Builder, keyword, description string, schemas openapi3.SchemaRefs, prefix string, indent, maxDepth int) {
	fmt.Fprintf(result, "%s- **%s** (%s):\n", prefix, keyword, description)
	for i, schemaRef := range schemas {
		fmt.Fprintf(result, "%s  - Option %d:\n", prefix, i+1)
		if schemaRef.Value != nil {
			result.WriteString(FormatSchema(schemaRef.Value, indent+2, maxDepth-1))
		}
	}
}

// formatObjectSchema formats an object type schema.
func formatObjectSchema(result *strings.Builder, schema *openapi3.Schema, prefix string, indent, maxDepth int) {
	fmt.Fprintf(result, "%s- Type: `object`\n", prefix)

	if schema.Nullable {
		fmt.Fprintf(result, "%s- Nullable: `true`\n", prefix)
	}

	if len(schema.Properties) == 0 {
		return
	}

	fmt.Fprintf(result, "%s- Properties:\n", prefix)

	// Build required map for O(1) lookup
	requiredMap := buildRequiredMap(schema.Required)

	// Sort properties for deterministic output
	propNames := getSortedPropertyNames(schema.Properties)

	for _, propName := range propNames {
		propRef := schema.Properties[propName]
		if propRef == nil || propRef.Value == nil {
			continue
		}

		prop := propRef.Value
		required := ""
		if requiredMap[propName] {
			required = MarkerRequired
		}
		deprecated := ""
		if prop.Deprecated {
			deprecated = MarkerDeprecated
		}

		fmt.Fprintf(result, "%s  - **%s**%s%s", prefix, propName, required, deprecated)
		if prop.Description != "" {
			fmt.Fprintf(result, ": %s\n", prop.Description)
		} else {
			result.WriteString("\n")
		}

		fmt.Fprintf(result, "%s    - Type: `%s`\n", prefix, FormatType(prop))

		if prop.Format != "" {
			fmt.Fprintf(result, "%s    - Format: `%s`\n", prefix, prop.Format)
		}
		if prop.Default != nil {
			fmt.Fprintf(result, "%s    - Default: `%v`\n", prefix, prop.Default)
		}
		if prop.Example != nil {
			fmt.Fprintf(result, "%s    - Example: `%v`\n", prefix, prop.Example)
		}
		if prop.Nullable {
			fmt.Fprintf(result, "%s    - Nullable: `true`\n", prefix)
		}

		constraints := FormatConstraints(prop)
		if constraints != "" {
			fmt.Fprintf(result, "%s    - Constraints: %s\n", prefix, constraints)
		}

		if len(prop.Enum) > 0 {
			fmt.Fprintf(result, "%s    - Allowed values: %v\n", prefix, prop.Enum)
		}

		// Recurse for nested objects and arrays
		if prop.Type.Is("object") && len(prop.Properties) > 0 {
			result.WriteString(FormatSchema(prop, indent+2, maxDepth-1))
		}
		if prop.Type.Is("array") && prop.Items != nil && prop.Items.Value != nil {
			fmt.Fprintf(result, "%s    - Items:\n", prefix)
			result.WriteString(FormatSchema(prop.Items.Value, indent+3, maxDepth-1))
		}
	}
}

// formatArraySchema formats an array type schema.
func formatArraySchema(result *strings.Builder, schema *openapi3.Schema, prefix string, indent, maxDepth int) {
	fmt.Fprintf(result, "%s- Type: `array`\n", prefix)

	if schema.Nullable {
		fmt.Fprintf(result, "%s- Nullable: `true`\n", prefix)
	}

	constraints := FormatConstraints(schema)
	if constraints != "" {
		fmt.Fprintf(result, "%s- Constraints: %s\n", prefix, constraints)
	}

	if schema.Items != nil && schema.Items.Value != nil {
		fmt.Fprintf(result, "%s- Items:\n", prefix)
		result.WriteString(FormatSchema(schema.Items.Value, indent+1, maxDepth-1))
	}
}

// formatPrimitiveSchema formats a primitive type schema (string, number, boolean, etc.).
func formatPrimitiveSchema(result *strings.Builder, schema *openapi3.Schema, prefix string) {
	fmt.Fprintf(result, "%s- Type: `%s`\n", prefix, FormatType(schema))

	if schema.Format != "" {
		fmt.Fprintf(result, "%s- Format: `%s`\n", prefix, schema.Format)
	}
	if schema.Nullable {
		fmt.Fprintf(result, "%s- Nullable: `true`\n", prefix)
	}
	if schema.Default != nil {
		fmt.Fprintf(result, "%s- Default: `%v`\n", prefix, schema.Default)
	}
	if schema.Example != nil {
		fmt.Fprintf(result, "%s- Example: `%v`\n", prefix, schema.Example)
	}

	constraints := FormatConstraints(schema)
	if constraints != "" {
		fmt.Fprintf(result, "%s- Constraints: %s\n", prefix, constraints)
	}

	if len(schema.Enum) > 0 {
		fmt.Fprintf(result, "%s- Allowed values: %v\n", prefix, schema.Enum)
	}
}
