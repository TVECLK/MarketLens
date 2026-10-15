package service

import (
	"context"
	"errors"
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
		if get(item) == "" {
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
	if name == "" {
		return ValidationErrorf("name is required")
	}
	if code == "" {
		return ValidationErrorf("code is required")
	}
	return nil
}
