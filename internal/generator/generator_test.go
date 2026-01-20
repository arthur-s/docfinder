package generator

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGenerateMarkdown_MultipleMethodsNoFilter(t *testing.T) {
	// Create a minimal OpenAPI document with an endpoint that has GET, PUT, DELETE
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Servers: []*openapi3.Server{
			{
				URL:         "https://api.example.com",
				Description: "Test Server",
			},
		},
	}

	// Create path item with multiple methods
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "Get item",
			Description: "Retrieves an item by ID",
			OperationID: "getItem",
			Tags:        []string{"Items"},
		},
		Put: &openapi3.Operation{
			Summary:     "Update item",
			Description: "Updates an existing item",
			OperationID: "updateItem",
			Tags:        []string{"Items"},
		},
		Delete: &openapi3.Operation{
			Summary:     "Delete item",
			Description: "Deletes an item by ID",
			OperationID: "deleteItem",
			Tags:        []string{"Items"},
		},
	}

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items/{id}", pathItem, "")

	// Verify all three methods appear in output
	if !strings.Contains(markdown, "## DELETE /items/{id}") {
		t.Error("Expected DELETE method in output when no filter specified")
	}
	if !strings.Contains(markdown, "## GET /items/{id}") {
		t.Error("Expected GET method in output when no filter specified")
	}
	if !strings.Contains(markdown, "## PUT /items/{id}") {
		t.Error("Expected PUT method in output when no filter specified")
	}

	// Verify operation details are present
	if !strings.Contains(markdown, "Get item") {
		t.Error("Expected GET operation summary in output")
	}
	if !strings.Contains(markdown, "Update item") {
		t.Error("Expected PUT operation summary in output")
	}
	if !strings.Contains(markdown, "Delete item") {
		t.Error("Expected DELETE operation summary in output")
	}
}

func TestGenerateMarkdown_MethodFilterGET(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "Get item",
			Description: "Retrieves an item by ID",
			OperationID: "getItem",
		},
		Put: &openapi3.Operation{
			Summary:     "Update item",
			Description: "Updates an existing item",
			OperationID: "updateItem",
		},
		Delete: &openapi3.Operation{
			Summary:     "Delete item",
			Description: "Deletes an item by ID",
			OperationID: "deleteItem",
		},
	}

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items/{id}", pathItem, "GET")

	// Verify only GET method appears
	if !strings.Contains(markdown, "## GET /items/{id}") {
		t.Error("Expected GET method in output when filtering by GET")
	}
	if strings.Contains(markdown, "## PUT /items/{id}") {
		t.Error("Did not expect PUT method in output when filtering by GET")
	}
	if strings.Contains(markdown, "## DELETE /items/{id}") {
		t.Error("Did not expect DELETE method in output when filtering by GET")
	}

	// Verify only GET operation details are present
	if !strings.Contains(markdown, "Get item") {
		t.Error("Expected GET operation summary in output")
	}
	if strings.Contains(markdown, "Update item") {
		t.Error("Did not expect PUT operation summary when filtering by GET")
	}
	if strings.Contains(markdown, "Delete item") {
		t.Error("Did not expect DELETE operation summary when filtering by GET")
	}
}

func TestGenerateMarkdown_MethodFilterPUT(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

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

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items/{id}", pathItem, "PUT")

	// Verify only PUT method appears
	if strings.Contains(markdown, "## GET /items/{id}") {
		t.Error("Did not expect GET method in output when filtering by PUT")
	}
	if !strings.Contains(markdown, "## PUT /items/{id}") {
		t.Error("Expected PUT method in output when filtering by PUT")
	}
	if strings.Contains(markdown, "## DELETE /items/{id}") {
		t.Error("Did not expect DELETE method in output when filtering by PUT")
	}
}

func TestGenerateMarkdown_MethodFilterDELETE(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

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

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items/{id}", pathItem, "DELETE")

	// Verify only DELETE method appears
	if strings.Contains(markdown, "## GET /items/{id}") {
		t.Error("Did not expect GET method in output when filtering by DELETE")
	}
	if strings.Contains(markdown, "## PUT /items/{id}") {
		t.Error("Did not expect PUT method in output when filtering by DELETE")
	}
	if !strings.Contains(markdown, "## DELETE /items/{id}") {
		t.Error("Expected DELETE method in output when filtering by DELETE")
	}
}

func TestGenerateMarkdown_SingleMethodEndpoint(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// Endpoint with only POST method
	pathItem := &openapi3.PathItem{
		Post: &openapi3.Operation{
			Summary:     "Create item",
			Description: "Creates a new item",
			OperationID: "createItem",
		},
	}

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items", pathItem, "")

	// Verify only POST appears
	if !strings.Contains(markdown, "## POST /items") {
		t.Error("Expected POST method in output")
	}
	if strings.Contains(markdown, "## GET /items") {
		t.Error("Did not expect GET method for POST-only endpoint")
	}
	if strings.Contains(markdown, "Create item") {
		// Good - summary is present
	} else {
		t.Error("Expected operation summary in output")
	}
}

func TestGenerateMarkdown_NilPathItem(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items", nil, "")

	if markdown != "" {
		t.Error("Expected empty string for nil pathItem")
	}
}

func TestGenerateMarkdown_EmptyPathItem(t *testing.T) {
	doc := &openapi3.T{
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}

	// PathItem with no operations
	pathItem := &openapi3.PathItem{}

	gen := New(doc)
	markdown := gen.GenerateMarkdown("/items", pathItem, "")

	// Should generate header but no operations
	if !strings.Contains(markdown, "# API Endpoint: /items") {
		t.Error("Expected endpoint header in output")
	}
	if strings.Contains(markdown, "## GET") || strings.Contains(markdown, "## POST") {
		t.Error("Did not expect any operation headers for empty pathItem")
	}
}
