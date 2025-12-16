package publisher

import (
	"regexp"
	"strings"
)

// Variable represents a parsed variable from input
type Variable struct {
	Key   string
	Value string
}

// ParseVariables parses space-separated key=value pairs
// Example: "user=john env=prod timestamp=2024-01-01"
func ParseVariables(input string) []Variable {
	var vars []Variable

	if strings.TrimSpace(input) == "" {
		return vars
	}

	// Split by spaces, but handle potential edge cases
	parts := strings.Fields(input)

	for _, part := range parts {
		// Find first = sign
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue // Skip invalid entries (no = or empty key)
		}

		key := part[:idx]
		value := part[idx+1:]

		// Validate key (alphanumeric and underscore only)
		if !isValidKey(key) {
			continue
		}

		vars = append(vars, Variable{
			Key:   key,
			Value: value,
		})
	}

	return vars
}

// isValidKey checks if a key contains only alphanumeric characters and underscores
func isValidKey(key string) bool {
	for _, r := range key {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return len(key) > 0
}

// SubstituteVariables replaces ${varName} placeholders with values
func SubstituteVariables(content string, vars []Variable) string {
	result := content

	for _, v := range vars {
		placeholder := "${" + v.Key + "}"
		result = strings.ReplaceAll(result, placeholder, v.Value)
	}

	return result
}

// FindVariables finds all ${varName} placeholders in content
func FindVariables(content string) []string {
	re := regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	var vars []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			vars = append(vars, match[1])
			seen[match[1]] = true
		}
	}

	return vars
}

// HasUnsubstitutedVariables checks if content still contains ${var} placeholders
func HasUnsubstitutedVariables(content string) bool {
	re := regexp.MustCompile(`\$\{[a-zA-Z_][a-zA-Z0-9_]*\}`)
	return re.MatchString(content)
}
