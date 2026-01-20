package generator

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FormatType returns a human-readable type string from an OpenAPI schema.
// Returns "unknown" if the schema is nil or has no type information.
// For multiple types, returns them joined with " | ".
func FormatType(schema *openapi3.Schema) string {
	if schema == nil {
		return "unknown"
	}

	types := schema.Type.Slice()
	if len(types) == 0 {
		return "unknown"
	}

	if len(types) == 1 {
		return types[0]
	}

	// Multiple types - join with pipe separator
	return strings.Join(types, " | ")
}

// FormatConstraints returns a comma-separated string of validation constraints
// for a schema (minLength, maxLength, pattern, min, max, etc.).
// Returns empty string if there are no constraints.
func FormatConstraints(schema *openapi3.Schema) string {
	if schema == nil {
		return ""
	}

	var constraints []string

	// String constraints
	if schema.MinLength > 0 {
		constraints = append(constraints, fmt.Sprintf("minLength: %d", schema.MinLength))
	}
	if schema.MaxLength != nil {
		constraints = append(constraints, fmt.Sprintf("maxLength: %d", *schema.MaxLength))
	}
	if schema.Pattern != "" {
		constraints = append(constraints, fmt.Sprintf("pattern: `%s`", schema.Pattern))
	}

	// Number constraints
	if schema.Min != nil {
		exclusive := ""
		if schema.ExclusiveMin {
			exclusive = " (exclusive)"
		}
		constraints = append(constraints, fmt.Sprintf("min: %v%s", *schema.Min, exclusive))
	}
	if schema.Max != nil {
		exclusive := ""
		if schema.ExclusiveMax {
			exclusive = " (exclusive)"
		}
		constraints = append(constraints, fmt.Sprintf("max: %v%s", *schema.Max, exclusive))
	}
	if schema.MultipleOf != nil {
		constraints = append(constraints, fmt.Sprintf("multipleOf: %v", *schema.MultipleOf))
	}

	// Array constraints
	if schema.MinItems > 0 {
		constraints = append(constraints, fmt.Sprintf("minItems: %d", schema.MinItems))
	}
	if schema.MaxItems != nil {
		constraints = append(constraints, fmt.Sprintf("maxItems: %d", *schema.MaxItems))
	}
	if schema.UniqueItems {
		constraints = append(constraints, "uniqueItems: true")
	}

	// Object constraints
	if schema.MinProps > 0 {
		constraints = append(constraints, fmt.Sprintf("minProperties: %d", schema.MinProps))
	}
	if schema.MaxProps != nil {
		constraints = append(constraints, fmt.Sprintf("maxProperties: %d", *schema.MaxProps))
	}

	if len(constraints) == 0 {
		return ""
	}

	return strings.Join(constraints, ", ")
}

// FormatJSON converts a value to pretty-printed JSON.
// Returns "{}" if value is nil.
// Returns the value formatted with %v if JSON marshaling fails.
func FormatJSON(value interface{}) (string, error) {
	if value == nil {
		return "{}", nil
	}

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// buildRequiredMap creates a map of required field names for O(1) lookup.
func buildRequiredMap(required []string) map[string]bool {
	requiredMap := make(map[string]bool, len(required))
	for _, req := range required {
		requiredMap[req] = true
	}
	return requiredMap
}

// getSortedKeys returns sorted keys from a map for deterministic iteration.
func getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// getSortedPropertyNames returns sorted property names from an OpenAPI schema.
func getSortedPropertyNames(properties openapi3.Schemas) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getSortedHeaderNames returns sorted header names from response headers.
func getSortedHeaderNames(headers openapi3.Headers) []string {
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getSortedExampleNames returns sorted example names.
func getSortedExampleNames(examples map[string]*openapi3.ExampleRef) []string {
	names := make([]string, 0, len(examples))
	for name := range examples {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getSortedContentTypes returns sorted content types from a content map.
func getSortedContentTypes(content openapi3.Content) []string {
	types := make([]string, 0, len(content))
	for ct := range content {
		types = append(types, ct)
	}
	sort.Strings(types)
	return types
}

// getSortedStatusCodes returns sorted status codes from responses.
func getSortedStatusCodes(responses map[string]*openapi3.ResponseRef) []string {
	codes := make([]string, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
