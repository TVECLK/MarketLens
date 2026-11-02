package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

// validateDateRange is the one shared "to must not be before from" rule
// used by every date-range query across the service package.
func validateDateRange(fromDate, toDate time.Time) error {
	if toDate.Before(fromDate) {
		return ValidationErrorf("to-date must not be before from-date")
	}
	return nil
}

// requireNonEmptyField builds a Create validator for the common flat-lookup
// shape: one required display string. get reads that field off the item
// (e.g. func(e *models.EmploymentSector) string { return e.Sector }).
func requireNonEmptyField[T any](fieldName string, get func(*T) string) func(context.Context, *T) error {
	return func(_ context.Context, item *T) error {
		if strings.TrimSpace(get(item)) == "" {
			return ValidationErrorf("%s is required", fieldName)
		}
		return nil
	}
}

// ParseDateRange parses a required from/to date pair (format YYYY-MM-DD) —
// the one place this parsing lives, called by both the HTTP handlers (via
// api.parseDateRangeQuery) and the MCP tools, instead of each transport
// reimplementing it. It only parses; the "to must not be before from" rule
// stays in validateDateRange, which callers that need it already run
// downstream.
func ParseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	if fromStr == "" {
		return time.Time{}, time.Time{}, ValidationErrorf("from-date is required (format YYYY-MM-DD)")
	}
	if toStr == "" {
		return time.Time{}, time.Time{}, ValidationErrorf("to-date is required (format YYYY-MM-DD)")
	}
	const layout = "2006-01-02"
	from, err := time.Parse(layout, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, ValidationErrorf("invalid from-date, expected YYYY-MM-DD: %v", err)
	}
	to, err := time.Parse(layout, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, ValidationErrorf("invalid to-date, expected YYYY-MM-DD: %v", err)
	}
	return from, to, nil
}

// validateParentExists turns a parent lookup's "not found" into a
// validation error naming the offending field; any other error (a real DB
// problem) passes through unchanged so it surfaces as a 500, not a
// misleading "doesn't exist". It takes the already-performed lookup's
// error rather than doing the lookup itself, since each entity's parent
// type differs.
func validateParentExists(lookupErr error, notFound error, field string, id uint) error {
	if errors.Is(lookupErr, notFound) {
		return ValidationErrorf("%s %d does not exist", field, id)
	}
	return lookupErr
}

// validateNameAndCode is the shared rule for every occupation/industry
// classification level: both fields are required.
func validateNameAndCode(name, code string) error {
	if strings.TrimSpace(name) == "" {
		return ValidationErrorf("name is required")
	}
	if strings.TrimSpace(code) == "" {
		return ValidationErrorf("code is required")
	}
	return nil
}

// validOccupationLevels / validIndustryLevels / leafLevels
var validOccupationLevels = map[string]bool{
	"major-group":      true,
	"sub-major-group":  true,
	"minor-group":      true,
	"unit-group":       true,
	"occupation-group": true,
}

var validIndustryLevels = map[string]bool{
	"industry-sector":   true,
	"industry-division": true,
	"industry-group":    true,
	"industry-class":    true,
	"industry-subclass": true,
}

// leafLevels marks the bottom of each hierarchy, which has no children to return.
var leafLevels = map[string]bool{
	"occupation-group":  true,
	"industry-subclass": true,
}

// validateStandard checks that standard is "occupation" or "industry", for
// endpoints scoped to a whole hierarchy rather than one specific level
// within it.
func validateStandard(standard string) error {
	if standard != "occupation" && standard != "industry" {
		return ValidationErrorf("invalid standard %q, must be 'occupation' or 'industry'", standard)
	}
	return nil
}

// validateStandardAndLevel additionally checks that level is valid for the
// given standard.
func validateStandardAndLevel(standard, level string) error {
	if err := validateStandard(standard); err != nil {
		return err
	}
	switch standard {
	case "occupation":
		if !validOccupationLevels[level] {
			return ValidationErrorf("invalid level %q for standard 'occupation'", level)
		}
	case "industry":
		if !validIndustryLevels[level] {
			return ValidationErrorf("invalid level %q for standard 'industry'", level)
		}
	}
	return nil
}
