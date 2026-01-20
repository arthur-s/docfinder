# DocFinder

A CLI tool for extracting and formatting documentation for specific API endpoints from OpenAPI 3.x specification files.

## Features

- ✅ **Multi-method support**: Filter by HTTP method (GET, POST, PUT, DELETE, etc.)
- ✅ **Short syntax**: `docfinder GET /books/{id} spec.yaml` (no flags needed!)
- ✅ **Smart detection**: Case-insensitive, auto-recognizes HTTP methods
- ✅ **Clear errors**: Shows available methods when invalid method specified
- ✅ **Comprehensive output**: Schema details, validation constraints, examples

## Installation

```bash
go build -o docfinder cmd/docfinder/main.go
```

## Usage

```bash
# Show all methods
docfinder /books/{book_id} openapi.yaml

# Filter by method (short syntax - recommended)
docfinder GET /books/{book_id} openapi.yaml
docfinder POST /books openapi.yaml
docfinder put /books/{book_id} openapi.yaml        # case-insensitive

# Alternative: use -method flag
docfinder -method DELETE /books/{book_id} openapi.yaml

# Generate docs to files
docfinder GET /books/{book_id} api.yaml > docs/get-book.md
docfinder POST /books api.yaml > docs/create-book.md

# Batch processing
for method in GET POST PUT DELETE; do
  docfinder $method /books/{book_id} api.yaml > docs/${method,,}-book.md
done
```

## Command-Line Help

```
Usage:
  docfinder [METHOD] <endpoint-path> <openapi-file>
  docfinder -method METHOD <endpoint-path> <openapi-file>

Examples:
  docfinder /books/{book_id} openapi.yaml                    # All methods
  docfinder GET /books/{book_id} openapi.yaml                # GET only
  docfinder -method DELETE /books/{book_id} openapi.yaml     # Flag syntax

Arguments:
  METHOD          Optional HTTP method (GET, POST, PUT, DELETE, PATCH, etc.)
  endpoint-path   API endpoint path to extract documentation for
  openapi-file    Path to OpenAPI YAML specification file

Flags:
  -method string  HTTP method to filter. If not specified, shows all methods.
```

## Output Format

Generated markdown includes:
- API metadata (title, version, base URLs)
- HTTP method and endpoint path
- Operation summary, description, and tags
- Parameters (path, query, header) with types and constraints
- Request/response body schemas with examples
- Security requirements
- Deprecation warnings

## Testing

```bash
go test ./...
```

## Contributing

Contributions welcome! Please ensure:
- All tests pass (`go test ./...`)
- Code is formatted with `gofmt`
- New features include tests
- Documentation is updated

## License

MIT
