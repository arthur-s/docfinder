package generator

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		schema   *openapi3.Schema
		expected string
	}{
		{
			name:     "nil schema",
			schema:   nil,
			expected: "unknown",
		},
		{
			name:     "empty schema",
			schema:   &openapi3.Schema{},
			expected: "unknown",
		},
		{
			name: "single type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
			expected: "string",
		},
		{
			name: "multiple types",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string", "null"},
			},
			expected: "string | null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatType(tt.schema)
			if result != tt.expected {
				t.Errorf("FormatType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatConstraints(t *testing.T) {
	minLen := uint64(5)
	maxLen := uint64(100)
	min := float64(0)
	max := float64(100)
	maxItems := uint64(10)
	maxProps := uint64(5)

	tests := []struct {
		name     string
		schema   *openapi3.Schema
		expected string
	}{
		{
			name:     "nil schema",
			schema:   nil,
			expected: "",
		},
		{
			name:     "no constraints",
			schema:   &openapi3.Schema{},
			expected: "",
		},
		{
			name: "string constraints",
			schema: &openapi3.Schema{
				MinLength: 5,
				MaxLength: &maxLen,
				Pattern:   "^[a-z]+$",
			},
			expected: "minLength: 5, maxLength: 100, pattern: `^[a-z]+$`",
		},
		{
			name: "number constraints",
			schema: &openapi3.Schema{
				Min: &min,
				Max: &max,
			},
			expected: "min: 0, max: 100",
		},
		{
			name: "exclusive minimum",
			schema: &openapi3.Schema{
				Min:          &min,
				ExclusiveMin: true,
			},
			expected: "min: 0 (exclusive)",
		},
		{
			name: "array constraints",
			schema: &openapi3.Schema{
				MinItems:    1,
				MaxItems:    &maxItems,
				UniqueItems: true,
			},
			expected: "minItems: 1, maxItems: 10, uniqueItems: true",
		},
		{
			name: "object constraints",
			schema: &openapi3.Schema{
				MinProps: 1,
				MaxProps: &maxProps,
			},
			expected: "minProperties: 1, maxProperties: 5",
		},
		{
			name: "mixed constraints",
			schema: &openapi3.Schema{
				MinLength: minLen,
				MaxLength: &maxLen,
				Pattern:   "^test$",
			},
			expected: "minLength: 5, maxLength: 100, pattern: `^test$`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConstraints(tt.schema)
			if result != tt.expected {
				t.Errorf("FormatConstraints() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectError bool
		contains    string
	}{
		{
			name:        "nil value",
			value:       nil,
			expectError: false,
			contains:    "{}",
		},
		{
			name: "simple object",
			value: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
			expectError: false,
			contains:    `"name": "test"`,
		},
		{
			name: "array",
			value: []string{
				"one", "two", "three",
			},
			expectError: false,
			contains:    `"one"`,
		},
		{
			name:        "string",
			value:       "hello",
			expectError: false,
			contains:    `"hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatJSON(tt.value)
			if (err != nil) != tt.expectError {
				t.Errorf("FormatJSON() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result == "" {
				t.Error("FormatJSON() returned empty string")
			}
			if tt.contains != "" && !contains(result, tt.contains) {
				t.Errorf("FormatJSON() = %v, should contain %v", result, tt.contains)
			}
		})
	}
}

func TestGetSortedPropertyNames(t *testing.T) {
	properties := openapi3.Schemas{
		"zebra":  &openapi3.SchemaRef{},
		"alpha":  &openapi3.SchemaRef{},
		"middle": &openapi3.SchemaRef{},
	}

	result := getSortedPropertyNames(properties)

	expected := []string{"alpha", "middle", "zebra"}
	if len(result) != len(expected) {
		t.Fatalf("getSortedPropertyNames() length = %d, want %d", len(result), len(expected))
	}

	for i, name := range result {
		if name != expected[i] {
			t.Errorf("getSortedPropertyNames()[%d] = %s, want %s", i, name, expected[i])
		}
	}
}

func TestGetSortedContentTypes(t *testing.T) {
	content := openapi3.Content{
		"text/plain":       &openapi3.MediaType{},
		"application/json": &openapi3.MediaType{},
		"application/xml":  &openapi3.MediaType{},
	}

	result := getSortedContentTypes(content)

	expected := []string{"application/json", "application/xml", "text/plain"}
	if len(result) != len(expected) {
		t.Fatalf("getSortedContentTypes() length = %d, want %d", len(result), len(expected))
	}

	for i, ct := range result {
		if ct != expected[i] {
			t.Errorf("getSortedContentTypes()[%d] = %s, want %s", i, ct, expected[i])
		}
	}
}

func TestBuildRequiredMap(t *testing.T) {
	required := []string{"id", "name", "email"}

	result := buildRequiredMap(required)

	if len(result) != 3 {
		t.Fatalf("buildRequiredMap() length = %d, want 3", len(result))
	}

	for _, field := range required {
		if !result[field] {
			t.Errorf("buildRequiredMap() missing field %s", field)
		}
	}

	if result["notpresent"] {
		t.Error("buildRequiredMap() should not contain 'notpresent'")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexStr(s, substr) >= 0)
}

func indexStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
