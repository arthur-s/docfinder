package generator

// Markdown heading constants
const (
	HeaderParameters  = "### Parameters\n\n"
	HeaderRequestBody = "### Request Body\n\n"
	HeaderResponses   = "### Responses\n\n"
	HeaderSecurity    = "### Security\n\n"
	HeaderExamples    = "\n**Examples:**\n\n"
	HeaderHeaders     = "**Headers:**\n\n"
	HeaderSchema      = "**Schema:**\n\n"

	SeparatorOperation = "---\n\n"
	MarkerRequired     = " **(required)**"
	MarkerDeprecated   = " ⚠️ *deprecated*"
)

// MaxRecursionDepth is the maximum depth for recursive schema formatting
// to prevent stack overflow on circular references or deeply nested schemas.
const MaxRecursionDepth = 20
