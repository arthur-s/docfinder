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

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <endpoint-path> <openapi-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s /app/v1/events/{id} openapi.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  endpoint-path   API endpoint path to extract documentation for\n")
		fmt.Fprintf(os.Stderr, "  openapi-file    Path to OpenAPI YAML specification file\n")
	}

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	endpointPath := flag.Arg(0)
	openapiFile := flag.Arg(1)

	if err := run(endpointPath, openapiFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(endpointPath, openapiFile string) error {
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

	// Generate markdown documentation
	gen := generator.New(doc)
	markdown := gen.GenerateMarkdown(endpointPath, pathItem)
	fmt.Print(markdown)

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
