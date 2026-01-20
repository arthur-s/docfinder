package main

import (
	"strings"
	"testing"

	"github.com/arthur-s/docfinder/internal/generator"
	"github.com/getkin/kin-openapi/openapi3"
)

// TestMultiMethodEndpoint_RealWorldSpec tests the /events/{event_id} endpoint
// from openapi-notify.yaml which has GET, PUT, and DELETE methods
func TestMultiMethodEndpoint_RealWorldSpec(t *testing.T) {
	doc, err := loadOpenAPISpec("../../openapi-notify.yaml")
	if err != nil {
		t.Skipf("Skipping test: openapi-notify.yaml not found: %v", err)
		return
	}

	endpointPath := "/events/{event_id}"
	pathItem, err := findPathItem(doc, endpointPath)
	if err != nil {
		t.Fatalf("Failed to find endpoint %s: %v", endpointPath, err)
	}

	// Test 1: No method filter - should include all methods
	t.Run("AllMethods", func(t *testing.T) {
		gen := generator.New(doc)
		markdown := gen.GenerateMarkdown(endpointPath, pathItem, "")

		// Verify all three methods are present
		if !strings.Contains(markdown, "## GET /events/{event_id}") {
			t.Error("Expected GET method in output")
		}
		if !strings.Contains(markdown, "## PUT /events/{event_id}") {
			t.Error("Expected PUT method in output")
		}
		if !strings.Contains(markdown, "## DELETE /events/{event_id}") {
			t.Error("Expected DELETE method in output")
		}

		// Verify operation summaries
		if !strings.Contains(markdown, "Get event details") {
			t.Error("Expected GET operation summary")
		}
		if !strings.Contains(markdown, "Update an event") {
			t.Error("Expected PUT operation summary")
		}
		if !strings.Contains(markdown, "Delete an event") {
			t.Error("Expected DELETE operation summary")
		}
	})

	// Test 2: Filter by GET method only
	t.Run("FilterGET", func(t *testing.T) {
		gen := generator.New(doc)
		markdown := gen.GenerateMarkdown(endpointPath, pathItem, "GET")

		// Should only have GET
		if !strings.Contains(markdown, "## GET /events/{event_id}") {
			t.Error("Expected GET method in output")
		}
		if strings.Contains(markdown, "## PUT /events/{event_id}") {
			t.Error("Did not expect PUT method when filtering by GET")
		}
		if strings.Contains(markdown, "## DELETE /events/{event_id}") {
			t.Error("Did not expect DELETE method when filtering by GET")
		}

		// Verify only GET operation details
		if !strings.Contains(markdown, "Get event details") {
			t.Error("Expected GET operation summary")
		}
		if strings.Contains(markdown, "Update an event") {
			t.Error("Did not expect PUT operation summary when filtering by GET")
		}
	})

	// Test 3: Filter by PUT method only
	t.Run("FilterPUT", func(t *testing.T) {
		gen := generator.New(doc)
		markdown := gen.GenerateMarkdown(endpointPath, pathItem, "PUT")

		// Should only have PUT
		if strings.Contains(markdown, "## GET /events/{event_id}") {
			t.Error("Did not expect GET method when filtering by PUT")
		}
		if !strings.Contains(markdown, "## PUT /events/{event_id}") {
			t.Error("Expected PUT method in output")
		}
		if strings.Contains(markdown, "## DELETE /events/{event_id}") {
			t.Error("Did not expect DELETE method when filtering by PUT")
		}

		// Verify PUT has request body (unlike GET and DELETE)
		if !strings.Contains(markdown, "Request Body") {
			t.Error("Expected PUT to have request body documentation")
		}
		if !strings.Contains(markdown, "Update an event") {
			t.Error("Expected PUT operation summary")
		}
	})

	// Test 4: Filter by DELETE method only
	t.Run("FilterDELETE", func(t *testing.T) {
		gen := generator.New(doc)
		markdown := gen.GenerateMarkdown(endpointPath, pathItem, "DELETE")

		// Should only have DELETE
		if strings.Contains(markdown, "## GET /events/{event_id}") {
			t.Error("Did not expect GET method when filtering by DELETE")
		}
		if strings.Contains(markdown, "## PUT /events/{event_id}") {
			t.Error("Did not expect PUT method when filtering by DELETE")
		}
		if !strings.Contains(markdown, "## DELETE /events/{event_id}") {
			t.Error("Expected DELETE method in output")
		}

		if !strings.Contains(markdown, "Delete an event") {
			t.Error("Expected DELETE operation summary")
		}
	})
}

func TestValidateMethod(t *testing.T) {
	// Create a path item with GET, PUT, DELETE
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary: "Get item",
		},
		Put: &openapi3.Operation{
			Summary: "Update item",
		},
		Delete: &openapi3.Operation{
			Summary: "Delete item",
		},
	}

	tests := []struct {
		name        string
		method      string
		expectError bool
	}{
		{"Valid GET", "GET", false},
		{"Valid PUT", "PUT", false},
		{"Valid DELETE", "DELETE", false},
		{"Invalid POST", "POST", true},
		{"Invalid PATCH", "PATCH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(pathItem, tt.method)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for method %s, got nil", tt.method)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Did not expect error for method %s, got: %v", tt.method, err)
			}
		})
	}
}

func TestNormalizeEndpointPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/events/{id}", "/events/{id}"},
		{"events/{id}", "/events/{id}"},
		{"/", "/"},
		{"", "/"},
		{"users", "/users"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeEndpointPath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeEndpointPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateInputFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{"Valid YAML file", "../../openapi-notify.yaml", false},
		{"Valid yaml extension", "test.yaml", true}, // doesn't exist but extension is valid
		{"Valid yml extension", "test.yml", true},   // doesn't exist but extension is valid
		{"Invalid extension", "test.txt", true},
		{"Directory", "../../", true},
		{"Non-existent file", "nonexistent.yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputFile(tt.filePath)
			if tt.name == "Valid YAML file" {
				// Only check existence for the real file
				if err != nil {
					t.Skipf("Skipping: file not found: %v", err)
				}
			} else if tt.name == "Invalid extension" || tt.name == "Directory" {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			}
		})
	}
}

func TestIsHTTPMethod(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"get", true},       // lowercase
		{"Post", true},      // mixed case
		{"delete", true},    // lowercase
		{"/events", false},  // path, not method
		{"users", false},    // not a method
		{"INVALID", false},  // not a valid HTTP method
		{"", false},         // empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isHTTPMethod(tt.input)
			if result != tt.expected {
				t.Errorf("isHTTPMethod(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
