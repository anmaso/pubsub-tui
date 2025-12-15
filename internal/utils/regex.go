package utils

import (
	"regexp"
)

// FilterResult contains the result of a regex filter operation
type FilterResult struct {
	Matches bool
	Error   error
}

// MatchesFilter checks if a string matches a regex pattern
// Returns true if the pattern is empty (no filter applied)
func MatchesFilter(text, pattern string) FilterResult {
	if pattern == "" {
		return FilterResult{Matches: true, Error: nil}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return FilterResult{Matches: false, Error: err}
	}

	return FilterResult{Matches: re.MatchString(text), Error: nil}
}

// ValidateRegex checks if a regex pattern is valid
func ValidateRegex(pattern string) error {
	if pattern == "" {
		return nil
	}
	_, err := regexp.Compile(pattern)
	return err
}

// FilterStrings filters a slice of strings by a regex pattern
// Returns the filtered slice and any regex compilation error
func FilterStrings(items []string, pattern string) ([]string, error) {
	if pattern == "" {
		return items, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return items, err
	}

	var filtered []string
	for _, item := range items {
		if re.MatchString(item) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

