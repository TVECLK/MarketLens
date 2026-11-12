package service

// Page-size rules for the paginated reads. They live here, in the service
// layer, so the REST handlers and the MCP tools share them instead of each
// transport keeping its own copy (a caller could otherwise pass any limit
// over MCP and pull a whole table).
const (
	// DefaultPageLimit and MaxPageLimit apply to the paginated list reads
	// (occupation groups, industry subclasses, skills by occupation).
	DefaultPageLimit = 20
	MaxPageLimit     = 100

	// DefaultHiringEmployersLimit is the default size of the top hiring
	// employers list; it is capped at MaxPageLimit.
	DefaultHiringEmployersLimit = 5
)

// pageLimit turns a caller-supplied limit into the one actually used: 0 or
// negative (not supplied) becomes def, and anything above max becomes max.
func pageLimit(requested, def, max int) int {
	if requested <= 0 {
		return def
	}
	if requested > max {
		return max
	}
	return requested
}

// pageOffset turns a negative offset into 0.
func pageOffset(requested int) int {
	if requested < 0 {
		return 0
	}
	return requested
}
