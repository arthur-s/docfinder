package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur-s/docfinder/internal/generator"
	"github.com/getkin/kin-openapi/openapi3"
)

const maxFileSize = 100 * 1024 * 1024 // 100MB limit

var (
	methodFlag = flag.String("method", "", "HTTP method to filter (GET, POST, PUT, DELETE, PATCH, etc.). If not specified, shows all methods.")
)

// Common HTTP methods for validation
var httpMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
	"TRACE":   true,
	"CONNECT": true,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [METHOD] <endpoint-path> <openapi-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -method METHOD <endpoint-path> <openapi-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s /events/{event_id} openapi.yaml                    # All methods\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s GET /events/{event_id} openapi.yaml                # GET only\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s PUT /events/{event_id} openapi.yaml                # PUT only\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -method DELETE /events/{event_id} openapi.yaml     # DELETE only\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  METHOD          Optional HTTP method (GET, POST, PUT, DELETE, PATCH, etc.)\n")
		fmt.Fprintf(os.Stderr, "  endpoint-path   API endpoint path to extract documentation for\n")
		fmt.Fprintf(os.Stderr, "  openapi-file    Path to OpenAPI YAML specification file\n")
	}

	flag.Parse()

	// Parse arguments - support both positional method and flag-based method
	var method, endpointPath, openapiFile string

	args := flag.Args()
	nArgs := len(args)

	// Case 1: 3 args - check if first arg is HTTP method (positional syntax)
	// Example: docfinder GET /events/{id} openapi.yaml
	if nArgs == 3 && isHTTPMethod(args[0]) {
		method = args[0]
		endpointPath = args[1]
		openapiFile = args[2]
	} else if nArgs == 2 {
		// Case 2: 2 args - standard format
		// Example: docfinder /events/{id} openapi.yaml
		// Or: docfinder -method GET /events/{id} openapi.yaml
		endpointPath = args[0]
		openapiFile = args[1]
		method = *methodFlag
	} else {
		flag.Usage()
		os.Exit(1)
	}

	// Flag takes precedence over positional method
	if *methodFlag != "" {
		method = *methodFlag
	}

	if err := run(endpointPath, openapiFile, method); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// isHTTPMethod checks if a string is a valid HTTP method
func isHTTPMethod(s string) bool {
	return httpMethods[strings.ToUpper(s)]
}

func run(endpointPath, openapiFile, method string) error {
	// Validate input file
	if err := validateInputFile(openapiFile); err != nil {
		return err
	}

	// Load OpenAPI specification
	doc, err := loadOpenAPISpec(openapiFile)
	if err != nil {
		return err
	}

	// Normalize the endpoint path (add leading slash if missing)
	endpointPath = normalizeEndpointPath(endpointPath)

	// Find the path item
	pathItem, err := findPathItem(doc, endpointPath)
	if err != nil {
		return err
	}

	// Normalize method (convert to uppercase for comparison with OpenAPI operations)
	method = strings.ToUpper(strings.TrimSpace(method))

	// Validate method if specified
	if method != "" {
		if err := validateMethod(pathItem, method); err != nil {
			return err
		}
	}

	// Generate markdown documentation
	gen := generator.New(doc)
	markdown := gen.GenerateMarkdown(endpointPath, pathItem, method)
	fmt.Print(markdown)

	return nil
}

// validateMethod checks if the specified HTTP method exists for the path item.
func validateMethod(pathItem *openapi3.PathItem, method string) error {
	operations := pathItem.Operations()

	// The operations map keys are already lowercase
	if operations[method] == nil {
		// Build a list of available methods (sorted for consistency)
		var available []string
		for m := range operations {
			available = append(available, m)
		}
		return fmt.Errorf("method '%s' not found for this endpoint. Available methods: %s",
			method, strings.Join(available, ", "))
	}
	return nil
}

// validateInputFile validates that the input file exists and is reasonable.
func validateInputFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	if info.Size() > maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return fmt.Errorf("unsupported file extension: %s (expected .yaml, .yml, or .json)", ext)
	}

	return nil
}

// loadOpenAPISpec loads and parses the OpenAPI specification file.
func loadOpenAPISpec(filePath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI file: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("loaded document is nil")
	}

	// Note: We skip validation because some OpenAPI files may have minor
	// spec violations but are still usable. We rely on the structure being
	// present rather than strict spec compliance.

	return doc, nil
}

// normalizeEndpointPath ensures the endpoint path starts with a slash.
func normalizeEndpointPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// findPathItem finds the path item for the given endpoint path.
func findPathItem(doc *openapi3.T, endpointPath string) (*openapi3.PathItem, error) {
	if doc.Paths == nil {
		return nil, fmt.Errorf("OpenAPI document has no paths defined")
	}

	pathItem := doc.Paths.Find(endpointPath)
	if pathItem == nil {
		return nil, fmt.Errorf("endpoint not found: %s", endpointPath)
	}

	return pathItem, nil
}
