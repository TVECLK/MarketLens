package models

import "strings"

// Standards are the two classification hierarchies.
const (
	StandardOccupation = "occupation"
	StandardIndustry   = "industry"
)

// hierarchies is the one ordered definition of each standard's levels,
// root first and leaf last. Valid levels, leaf levels, child levels, level
// table names and the MCP tool descriptions are all derived from it, so a
// level added or renamed here is picked up everywhere.
var hierarchies = map[string][]string{
	StandardOccupation: {"major-group", "sub-major-group", "minor-group", "unit-group", "occupation-group"},
	StandardIndustry:   {"industry-sector", "industry-division", "industry-group", "industry-class", "industry-subclass"},
}

// HierarchyLevels returns the levels of a standard, root first. It returns
// nil for an unknown standard. The result is a copy.
func HierarchyLevels(standard string) []string {
	return append([]string(nil), hierarchies[standard]...)
}

// IsValidLevel reports whether level belongs to the given standard.
func IsValidLevel(standard, level string) bool {
	for _, l := range hierarchies[standard] {
		if l == level {
			return true
		}
	}
	return false
}

// IsKnownLevel reports whether level belongs to either standard.
func IsKnownLevel(level string) bool {
	return IsValidLevel(StandardOccupation, level) || IsValidLevel(StandardIndustry, level)
}

// IsLeafLevel reports whether level is the last level of its hierarchy,
// which has no children.
func IsLeafLevel(level string) bool {
	for _, levels := range hierarchies {
		if levels[len(levels)-1] == level {
			return true
		}
	}
	return false
}

// ChildLevel returns the level directly below level, or false if level is
// a leaf or not part of the standard.
func ChildLevel(standard, level string) (string, bool) {
	levels := hierarchies[standard]
	for i, l := range levels {
		if l == level && i+1 < len(levels) {
			return levels[i+1], true
		}
	}
	return "", false
}

// LevelTable returns the database table backing a level ("major-group" ->
// "major_group"), or "" if level is not a known level, so an unvalidated
// string can never reach a query.
func LevelTable(level string) string {
	if !IsKnownLevel(level) {
		return ""
	}
	return strings.ReplaceAll(level, "-", "_")
}
