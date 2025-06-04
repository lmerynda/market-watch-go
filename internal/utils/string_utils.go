package utils

import (
	"strings"
)

// ParseSymbols parses a comma-separated string of symbols
func ParseSymbols(symbolsParam string) []string {
	if symbolsParam == "" {
		return []string{}
	}

	symbols := strings.Split(symbolsParam, ",")
	var result []string

	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol != "" {
			result = append(result, strings.ToUpper(symbol))
		}
	}

	return result
}

// BoolPtr returns a pointer to a boolean value
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an integer value
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// SafeDereference safely dereferences a pointer, returning the zero value if nil
func SafeDereference(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// SafeDereferenceString safely dereferences a string pointer
func SafeDereferenceString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// SafeDereferenceInt safely dereferences an int pointer
func SafeDereferenceInt(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// SafeDereferenceFloat64 safely dereferences a float64 pointer
func SafeDereferenceFloat64(ptr *float64) float64 {
	if ptr == nil {
		return 0.0
	}
	return *ptr
}
