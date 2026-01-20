# DocFinder

A CLI tool for extracting and formatting documentation for specific API endpoints from OpenAPI 3.x specification files.

## Features

- Extract documentation for a single endpoint from large OpenAPI files
- Output structured Markdown optimized for AI agents and human readers
- Comprehensive schema details including validation constraints
- Request and response examples with JSON formatting
- Support for complex schemas (oneOf, anyOf, allOf)
- Deterministic output (sorted properties, headers, examples)
- Response headers documentation
- Deprecation warnings

## Building

```bash
go build -o docfinder ./cmd/docfinder
```

## Usage

```bash
docfinder <endpoint-path> <openapi-file>
```

### Examples

```bash
# Extract documentation for a specific endpoint
docfinder /users openapi.yaml

# Works with parameterized paths
docfinder /posts/{post_id} openapi.yaml

# Leading slash is optional
docfinder users openapi.yaml
```

## Output Format

The tool generates comprehensive Markdown documentation including:

### Metadata
- API name and version
- Base URL(s) with descriptions
- Operation summary and description
- Operation ID and tags
- Deprecation warnings (if applicable)

### Parameters
- Name, location (path/query/header), and required status
- Type and format information
- Default values and examples
- Validation constraints (min/max length, pattern, etc.)
- Allowed values for enums

### Request Body
- Content types
- Detailed schema with nested properties
- Required/optional field indicators
- Type information with formats
- Validation constraints
- Request examples with JSON formatting

### Responses
- Status codes with descriptions
- Response headers (e.g., rate limiting headers)
- Content types
- Response schemas
- Response examples with JSON formatting

### Security
- Required authentication schemes
- OAuth scopes (if applicable)

## License

MIT

## Contributing

Contributions welcome! Please ensure:
1. All tests pass
2. Code is formatted with `gofmt`
3. New features include tests
4. Exported functions are documented
