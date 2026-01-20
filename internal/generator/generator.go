package generator

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Generator generates markdown documentation from OpenAPI specifications.
type Generator struct {
	doc *openapi3.T
}

// New creates a new Generator with the given OpenAPI document.
func New(doc *openapi3.T) *Generator {
	return &Generator{doc: doc}
}

// GenerateMarkdown generates markdown documentation for a specific endpoint.
// path is the endpoint path (e.g., "/users/{id}").
// pathItem contains the OpenAPI path item definition.
// method is an optional HTTP method filter (e.g., "GET", "POST"). Empty string means all methods.
// Returns a markdown-formatted string.
func (g *Generator) GenerateMarkdown(path string, pathItem *openapi3.PathItem, method string) string {
	if pathItem == nil {
		return ""
	}

	var md strings.Builder

	g.writeHeader(&md, path)
	g.writeOperations(&md, path, pathItem, method)

	return md.String()
}

// writeHeader writes the API metadata and server information.
func (g *Generator) writeHeader(md *strings.Builder, path string) {
	fmt.Fprintf(md, "# API Endpoint: %s\n\n", path)

	if g.doc.Info != nil {
		fmt.Fprintf(md, "**API:** %s %s\n\n", g.doc.Info.Title, g.doc.Info.Version)
	}

	// Server information
	if len(g.doc.Servers) > 0 {
		md.WriteString("**Base URL(s):**\n")
		for _, server := range g.doc.Servers {
			if server.Description != "" {
				fmt.Fprintf(md, "- `%s` - %s\n", server.URL, server.Description)
			} else {
				fmt.Fprintf(md, "- `%s`\n", server.URL)
			}
		}
		md.WriteString("\n")
	}
}

// writeOperations writes all HTTP operations for the endpoint, optionally filtered by method.
// methodFilter is an uppercase HTTP method (e.g., "GET", "POST") or empty string for all methods.
func (g *Generator) writeOperations(md *strings.Builder, path string, pathItem *openapi3.PathItem, methodFilter string) {
	for method, operation := range pathItem.Operations() {
		if operation == nil {
			continue
		}

		// Filter by method if specified
		if methodFilter != "" && method != methodFilter {
			continue
		}

		g.writeOperation(md, method, path, operation)
	}
}

// writeOperation writes a single HTTP operation.
func (g *Generator) writeOperation(md *strings.Builder, method, path string, operation *openapi3.Operation) {
	fmt.Fprintf(md, "## %s %s\n\n", strings.ToUpper(method), path)

	g.writeOperationMetadata(md, operation)
	g.writeParameters(md, operation.Parameters)
	g.writeRequestBody(md, operation.RequestBody)
	g.writeResponses(md, operation.Responses)
	g.writeSecurity(md, operation.Security)

	md.WriteString(SeparatorOperation)
}

// writeOperationMetadata writes operation summary, description, and tags.
func (g *Generator) writeOperationMetadata(md *strings.Builder, operation *openapi3.Operation) {
	// Deprecation warning
	if operation.Deprecated {
		md.WriteString("⚠️ **DEPRECATED** - This operation is deprecated and may be removed in a future version.\n\n")
	}

	if operation.Summary != "" {
		fmt.Fprintf(md, "**Summary:** %s\n\n", operation.Summary)
	}

	if operation.Description != "" {
		fmt.Fprintf(md, "**Description:** %s\n\n", operation.Description)
	}

	if operation.OperationID != "" {
		fmt.Fprintf(md, "**Operation ID:** `%s`\n\n", operation.OperationID)
	}

	if len(operation.Tags) > 0 {
		fmt.Fprintf(md, "**Tags:** %s\n\n", strings.Join(operation.Tags, ", "))
	}
}

// writeParameters writes parameter documentation.
func (g *Generator) writeParameters(md *strings.Builder, parameters openapi3.Parameters) {
	if len(parameters) == 0 {
		return
	}

	md.WriteString(HeaderParameters)

	for _, paramRef := range parameters {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}

		param := paramRef.Value
		required := ""
		if param.Required {
			required = MarkerRequired
		}
		deprecated := ""
		if param.Deprecated {
			deprecated = MarkerDeprecated
		}

		fmt.Fprintf(md, "- **%s** (%s)%s%s\n", param.Name, param.In, required, deprecated)

		if param.Description != "" {
			fmt.Fprintf(md, "  - Description: %s\n", param.Description)
		}

		if param.Schema != nil && param.Schema.Value != nil {
			schema := param.Schema.Value
			fmt.Fprintf(md, "  - Type: `%s`\n", FormatType(schema))

			if schema.Format != "" {
				fmt.Fprintf(md, "  - Format: `%s`\n", schema.Format)
			}
			if schema.Default != nil {
				fmt.Fprintf(md, "  - Default: `%v`\n", schema.Default)
			}
			if schema.Example != nil {
				fmt.Fprintf(md, "  - Example: `%v`\n", schema.Example)
			}

			constraints := FormatConstraints(schema)
			if constraints != "" {
				fmt.Fprintf(md, "  - Constraints: %s\n", constraints)
			}

			if len(schema.Enum) > 0 {
				fmt.Fprintf(md, "  - Allowed values: %v\n", schema.Enum)
			}
		}
	}

	md.WriteString("\n")
}

// writeRequestBody writes request body documentation.
func (g *Generator) writeRequestBody(md *strings.Builder, requestBodyRef *openapi3.RequestBodyRef) {
	if requestBodyRef == nil || requestBodyRef.Value == nil {
		return
	}

	reqBody := requestBodyRef.Value
	md.WriteString(HeaderRequestBody)

	if reqBody.Description != "" {
		fmt.Fprintf(md, "%s\n\n", reqBody.Description)
	}

	if reqBody.Required {
		md.WriteString("**Required:** (required)\n\n")
	} else {
		md.WriteString("**Required:** (optional)\n\n")
	}

	// Sort content types for deterministic output
	contentTypes := getSortedContentTypes(reqBody.Content)

	for _, contentType := range contentTypes {
		mediaType := reqBody.Content[contentType]
		if mediaType == nil {
			continue
		}

		fmt.Fprintf(md, "**Content-Type:** `%s`\n\n", contentType)

		if mediaType.Schema != nil && mediaType.Schema.Value != nil {
			md.WriteString(HeaderSchema)
			md.WriteString(FormatSchema(mediaType.Schema.Value, 0, MaxRecursionDepth))
		}

		g.writeExamples(md, mediaType.Examples)
	}

	md.WriteString("\n")
}

// writeResponses writes response documentation.
func (g *Generator) writeResponses(md *strings.Builder, responses *openapi3.Responses) {
	if responses == nil || responses.Map() == nil || len(responses.Map()) == 0 {
		return
	}

	md.WriteString(HeaderResponses)

	// Sort status codes for deterministic output
	statusCodes := getSortedStatusCodes(responses.Map())

	for _, status := range statusCodes {
		respRef := responses.Map()[status]
		if respRef == nil || respRef.Value == nil {
			continue
		}

		resp := respRef.Value
		fmt.Fprintf(md, "#### %s\n\n", status)

		if resp.Description != nil {
			fmt.Fprintf(md, "%s\n\n", *resp.Description)
		}

		g.writeResponseHeaders(md, resp.Headers)

		// Sort content types for deterministic output
		contentTypes := getSortedContentTypes(resp.Content)

		for _, contentType := range contentTypes {
			mediaType := resp.Content[contentType]
			if mediaType == nil {
				continue
			}

			fmt.Fprintf(md, "**Content-Type:** `%s`\n\n", contentType)

			if mediaType.Schema != nil && mediaType.Schema.Value != nil {
				md.WriteString(HeaderSchema)
				md.WriteString(FormatSchema(mediaType.Schema.Value, 0, MaxRecursionDepth))
			}

			g.writeExamples(md, mediaType.Examples)
		}

		md.WriteString("\n")
	}
}

// writeResponseHeaders writes response header documentation.
func (g *Generator) writeResponseHeaders(md *strings.Builder, headers openapi3.Headers) {
	if len(headers) == 0 {
		return
	}

	md.WriteString(HeaderHeaders)

	// Sort header names for deterministic output
	headerNames := getSortedHeaderNames(headers)

	for _, headerName := range headerNames {
		headerRef := headers[headerName]
		if headerRef == nil || headerRef.Value == nil {
			continue
		}

		header := headerRef.Value
		desc := ""
		if header.Description != "" {
			desc = fmt.Sprintf(" - %s", header.Description)
		}

		fmt.Fprintf(md, "- `%s`%s\n", headerName, desc)

		if header.Schema != nil && header.Schema.Value != nil {
			fmt.Fprintf(md, "  - Type: `%s`\n", FormatType(header.Schema.Value))
		}
	}

	md.WriteString("\n")
}

// writeExamples writes example documentation.
func (g *Generator) writeExamples(md *strings.Builder, examples map[string]*openapi3.ExampleRef) {
	if len(examples) == 0 {
		return
	}

	md.WriteString(HeaderExamples)

	// Sort example names for deterministic output
	exampleNames := getSortedExampleNames(examples)

	for _, exampleName := range exampleNames {
		exampleRef := examples[exampleName]
		if exampleRef == nil || exampleRef.Value == nil {
			continue
		}

		example := exampleRef.Value

		if example.Summary != "" {
			fmt.Fprintf(md, "*%s* (`%s`):\n\n", example.Summary, exampleName)
		} else {
			fmt.Fprintf(md, "*Example: `%s`*:\n\n", exampleName)
		}

		jsonStr, err := FormatJSON(example.Value)
		if err != nil {
			// Fallback to %v formatting if JSON marshal fails
			fmt.Fprintf(md, "```\n%v\n```\n\n", example.Value)
		} else {
			fmt.Fprintf(md, "```json\n%s\n```\n\n", jsonStr)
		}
	}
}

// writeSecurity writes security requirement documentation.
func (g *Generator) writeSecurity(md *strings.Builder, security *openapi3.SecurityRequirements) {
	if security == nil || len(*security) == 0 {
		return
	}

	md.WriteString(HeaderSecurity)

	for _, secReq := range *security {
		for name, scopes := range secReq {
			if len(scopes) > 0 {
				fmt.Fprintf(md, "- **%s**: %s\n", name, strings.Join(scopes, ", "))
			} else {
				fmt.Fprintf(md, "- **%s**\n", name)
			}
		}
	}

	md.WriteString("\n")
}
