package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

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
	// Load OpenAPI specification
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(openapiFile)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI file: %w", err)
	}

	// Skip validation - some OpenAPI files may have minor issues but are still usable
	// We'll rely on the structure being present rather than strict validation

	// Normalize the endpoint path (remove leading slash if missing)
	if !strings.HasPrefix(endpointPath, "/") {
		endpointPath = "/" + endpointPath
	}

	// Find the path item
	pathItem := doc.Paths.Find(endpointPath)
	if pathItem == nil {
		return fmt.Errorf("endpoint not found: %s", endpointPath)
	}

	// Generate markdown documentation
	markdown := generateMarkdown(doc, endpointPath, pathItem)
	fmt.Println(markdown)

	return nil
}

func generateMarkdown(doc *openapi3.T, path string, pathItem *openapi3.PathItem) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# API Endpoint: %s\n\n", path))

	if doc.Info != nil {
		md.WriteString(fmt.Sprintf("**API:** %s %s\n\n", doc.Info.Title, doc.Info.Version))
	}

	// Server information
	if len(doc.Servers) > 0 {
		md.WriteString("**Base URL(s):**\n")
		for _, server := range doc.Servers {
			if server.Description != "" {
				md.WriteString(fmt.Sprintf("- `%s` - %s\n", server.URL, server.Description))
			} else {
				md.WriteString(fmt.Sprintf("- `%s`\n", server.URL))
			}
		}
		md.WriteString("\n")
	}

	// Iterate through all operations
	for method, operation := range pathItem.Operations() {
		md.WriteString(fmt.Sprintf("## %s %s\n\n", strings.ToUpper(method), path))

		// Deprecation warning
		if operation.Deprecated {
			md.WriteString("⚠️ **DEPRECATED** - This operation is deprecated and may be removed in a future version.\n\n")
		}

		if operation.Summary != "" {
			md.WriteString(fmt.Sprintf("**Summary:** %s\n\n", operation.Summary))
		}

		if operation.Description != "" {
			md.WriteString(fmt.Sprintf("**Description:** %s\n\n", operation.Description))
		}

		if operation.OperationID != "" {
			md.WriteString(fmt.Sprintf("**Operation ID:** `%s`\n\n", operation.OperationID))
		}

		// Tags
		if len(operation.Tags) > 0 {
			md.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(operation.Tags, ", ")))
		}

		// Parameters
		if len(operation.Parameters) > 0 {
			md.WriteString("### Parameters\n\n")
			for _, paramRef := range operation.Parameters {
				param := paramRef.Value
				required := ""
				if param.Required {
					required = " **(required)**"
				}
				deprecated := ""
				if param.Deprecated {
					deprecated = " ⚠️ *deprecated*"
				}
				md.WriteString(fmt.Sprintf("- **%s** (%s)%s%s\n", param.Name, param.In, required, deprecated))
				if param.Description != "" {
					md.WriteString(fmt.Sprintf("  - Description: %s\n", param.Description))
				}
				if param.Schema != nil && param.Schema.Value != nil {
					schema := param.Schema.Value
					md.WriteString(fmt.Sprintf("  - Type: `%s`\n", formatType(schema)))
					if schema.Format != "" {
						md.WriteString(fmt.Sprintf("  - Format: `%s`\n", schema.Format))
					}
					if schema.Default != nil {
						md.WriteString(fmt.Sprintf("  - Default: `%v`\n", schema.Default))
					}
					if schema.Example != nil {
						md.WriteString(fmt.Sprintf("  - Example: `%v`\n", schema.Example))
					}
					// Validation constraints
					constraints := formatConstraints(schema)
					if constraints != "" {
						md.WriteString(fmt.Sprintf("  - Constraints: %s\n", constraints))
					}
					if len(schema.Enum) > 0 {
						md.WriteString(fmt.Sprintf("  - Allowed values: %v\n", schema.Enum))
					}
				}
			}
			md.WriteString("\n")
		}

		// Request Body
		if operation.RequestBody != nil && operation.RequestBody.Value != nil {
			md.WriteString("### Request Body\n\n")
			reqBody := operation.RequestBody.Value
			if reqBody.Description != "" {
				md.WriteString(fmt.Sprintf("%s\n\n", reqBody.Description))
			}
			required := ""
			if reqBody.Required {
				required = " (required)"
			}
			md.WriteString(fmt.Sprintf("**Required:** %s\n\n", strings.TrimSpace(required)))

			for contentType, mediaType := range reqBody.Content {
				md.WriteString(fmt.Sprintf("**Content-Type:** `%s`\n\n", contentType))
				if mediaType.Schema != nil && mediaType.Schema.Value != nil {
					md.WriteString("**Schema:**\n\n")
					md.WriteString(formatSchema(mediaType.Schema.Value, 0))
				}

				// Add examples
				if len(mediaType.Examples) > 0 {
					md.WriteString("\n**Examples:**\n\n")
					for exampleName, exampleRef := range mediaType.Examples {
						if exampleRef.Value != nil {
							if exampleRef.Value.Summary != "" {
								md.WriteString(fmt.Sprintf("*%s* (`%s`):\n\n", exampleRef.Value.Summary, exampleName))
							} else {
								md.WriteString(fmt.Sprintf("*Example: `%s`*:\n\n", exampleName))
							}
							md.WriteString("```json\n")
							md.WriteString(formatJSON(exampleRef.Value.Value))
							md.WriteString("\n```\n\n")
						}
					}
				}
			}
			md.WriteString("\n")
		}

		// Responses
		if operation.Responses != nil && operation.Responses.Map() != nil {
			md.WriteString("### Responses\n\n")
			for status, respRef := range operation.Responses.Map() {
				resp := respRef.Value
				md.WriteString(fmt.Sprintf("#### %s\n\n", status))
				if resp.Description != nil {
					md.WriteString(fmt.Sprintf("%s\n\n", *resp.Description))
				}

				// Response headers
				if len(resp.Headers) > 0 {
					md.WriteString("**Headers:**\n\n")
					for headerName, headerRef := range resp.Headers {
						if headerRef.Value != nil {
							header := headerRef.Value
							desc := ""
							if header.Description != "" {
								desc = fmt.Sprintf(" - %s", header.Description)
							}
							md.WriteString(fmt.Sprintf("- `%s`%s\n", headerName, desc))
							if header.Schema != nil && header.Schema.Value != nil {
								md.WriteString(fmt.Sprintf("  - Type: `%s`\n", formatType(header.Schema.Value)))
							}
						}
					}
					md.WriteString("\n")
				}

				for contentType, mediaType := range resp.Content {
					md.WriteString(fmt.Sprintf("**Content-Type:** `%s`\n\n", contentType))
					if mediaType.Schema != nil && mediaType.Schema.Value != nil {
						md.WriteString("**Schema:**\n\n")
						md.WriteString(formatSchema(mediaType.Schema.Value, 0))
					}

					// Add examples
					if len(mediaType.Examples) > 0 {
						md.WriteString("\n**Examples:**\n\n")
						for exampleName, exampleRef := range mediaType.Examples {
							if exampleRef.Value != nil {
								if exampleRef.Value.Summary != "" {
									md.WriteString(fmt.Sprintf("*%s* (`%s`):\n\n", exampleRef.Value.Summary, exampleName))
								} else {
									md.WriteString(fmt.Sprintf("*Example: `%s`*:\n\n", exampleName))
								}
								md.WriteString("```json\n")
								md.WriteString(formatJSON(exampleRef.Value.Value))
								md.WriteString("\n```\n\n")
							}
						}
					}
				}
				md.WriteString("\n")
			}
		}

		// Security
		if operation.Security != nil && len(*operation.Security) > 0 {
			md.WriteString("### Security\n\n")
			for _, secReq := range *operation.Security {
				for name, scopes := range secReq {
					if len(scopes) > 0 {
						md.WriteString(fmt.Sprintf("- **%s**: %s\n", name, strings.Join(scopes, ", ")))
					} else {
						md.WriteString(fmt.Sprintf("- **%s**\n", name))
					}
				}
			}
			md.WriteString("\n")
		}

		md.WriteString("---\n\n")
	}

	return md.String()
}

func formatSchema(schema *openapi3.Schema, indent int) string {
	var result strings.Builder
	prefix := strings.Repeat("  ", indent)

	// Handle oneOf, anyOf, allOf
	if len(schema.OneOf) > 0 {
		result.WriteString(fmt.Sprintf("%s- **oneOf** (one of the following):\n", prefix))
		for i, schemaRef := range schema.OneOf {
			result.WriteString(fmt.Sprintf("%s  - Option %d:\n", prefix, i+1))
			if schemaRef.Value != nil {
				result.WriteString(formatSchema(schemaRef.Value, indent+2))
			}
		}
		return result.String()
	}

	if len(schema.AnyOf) > 0 {
		result.WriteString(fmt.Sprintf("%s- **anyOf** (any of the following):\n", prefix))
		for i, schemaRef := range schema.AnyOf {
			result.WriteString(fmt.Sprintf("%s  - Option %d:\n", prefix, i+1))
			if schemaRef.Value != nil {
				result.WriteString(formatSchema(schemaRef.Value, indent+2))
			}
		}
		return result.String()
	}

	if len(schema.AllOf) > 0 {
		result.WriteString(fmt.Sprintf("%s- **allOf** (all of the following):\n", prefix))
		for i, schemaRef := range schema.AllOf {
			result.WriteString(fmt.Sprintf("%s  - Schema %d:\n", prefix, i+1))
			if schemaRef.Value != nil {
				result.WriteString(formatSchema(schemaRef.Value, indent+2))
			}
		}
		return result.String()
	}

	if schema.Type.Is("object") {
		result.WriteString(fmt.Sprintf("%s- Type: `object`\n", prefix))

		if schema.Nullable {
			result.WriteString(fmt.Sprintf("%s- Nullable: `true`\n", prefix))
		}

		if len(schema.Properties) > 0 {
			result.WriteString(fmt.Sprintf("%s- Properties:\n", prefix))
			for propName, propRef := range schema.Properties {
				prop := propRef.Value
				required := ""
				for _, req := range schema.Required {
					if req == propName {
						required = " **(required)**"
						break
					}
				}
				deprecated := ""
				if prop.Deprecated {
					deprecated = " ⚠️ *deprecated*"
				}

				result.WriteString(fmt.Sprintf("%s  - **%s**%s%s", prefix, propName, required, deprecated))
				if prop.Description != "" {
					result.WriteString(fmt.Sprintf(": %s\n", prop.Description))
				} else {
					result.WriteString("\n")
				}

				result.WriteString(fmt.Sprintf("%s    - Type: `%s`\n", prefix, formatType(prop)))

				if prop.Format != "" {
					result.WriteString(fmt.Sprintf("%s    - Format: `%s`\n", prefix, prop.Format))
				}
				if prop.Default != nil {
					result.WriteString(fmt.Sprintf("%s    - Default: `%v`\n", prefix, prop.Default))
				}
				if prop.Example != nil {
					result.WriteString(fmt.Sprintf("%s    - Example: `%v`\n", prefix, prop.Example))
				}
				if prop.Nullable {
					result.WriteString(fmt.Sprintf("%s    - Nullable: `true`\n", prefix))
				}

				constraints := formatConstraints(prop)
				if constraints != "" {
					result.WriteString(fmt.Sprintf("%s    - Constraints: %s\n", prefix, constraints))
				}

				if len(prop.Enum) > 0 {
					result.WriteString(fmt.Sprintf("%s    - Allowed values: %v\n", prefix, prop.Enum))
				}

				if prop.Type.Is("object") && len(prop.Properties) > 0 {
					result.WriteString(formatSchema(prop, indent+2))
				}
				if prop.Type.Is("array") && prop.Items != nil && prop.Items.Value != nil {
					result.WriteString(fmt.Sprintf("%s    - Items:\n", prefix))
					result.WriteString(formatSchema(prop.Items.Value, indent+3))
				}
			}
		}
	} else if schema.Type.Is("array") {
		result.WriteString(fmt.Sprintf("%s- Type: `array`\n", prefix))

		if schema.Nullable {
			result.WriteString(fmt.Sprintf("%s- Nullable: `true`\n", prefix))
		}

		constraints := formatConstraints(schema)
		if constraints != "" {
			result.WriteString(fmt.Sprintf("%s- Constraints: %s\n", prefix, constraints))
		}

		if schema.Items != nil && schema.Items.Value != nil {
			result.WriteString(fmt.Sprintf("%s- Items:\n", prefix))
			result.WriteString(formatSchema(schema.Items.Value, indent+1))
		}
	} else if schema.Type.Slice() != nil {
		result.WriteString(fmt.Sprintf("%s- Type: `%s`\n", prefix, formatType(schema)))

		if schema.Format != "" {
			result.WriteString(fmt.Sprintf("%s- Format: `%s`\n", prefix, schema.Format))
		}
		if schema.Nullable {
			result.WriteString(fmt.Sprintf("%s- Nullable: `true`\n", prefix))
		}
		if schema.Default != nil {
			result.WriteString(fmt.Sprintf("%s- Default: `%v`\n", prefix, schema.Default))
		}
		if schema.Example != nil {
			result.WriteString(fmt.Sprintf("%s- Example: `%v`\n", prefix, schema.Example))
		}

		constraints := formatConstraints(schema)
		if constraints != "" {
			result.WriteString(fmt.Sprintf("%s- Constraints: %s\n", prefix, constraints))
		}
	}

	if len(schema.Enum) > 0 && !schema.Type.Is("object") && !schema.Type.Is("array") {
		result.WriteString(fmt.Sprintf("%s- Allowed values: %v\n", prefix, schema.Enum))
	}

	return result.String()
}

func formatJSON(value interface{}) string {
	if value == nil {
		return "{}"
	}

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	return string(jsonBytes)
}

func formatType(schema *openapi3.Schema) string {
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

	// Multiple types
	return strings.Join(types, " | ")
}

func formatConstraints(schema *openapi3.Schema) string {
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
