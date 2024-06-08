package sim

import (
	"regexp"
	"strings"
)

// MatchWildcard determines if the provided string matches the wildcard pattern, using AWS's
// heuristics for wildcards
//
// TODO(nsiow) consider moving this to its own package
// TODO(nsiow) add trace logging for better debugging
// TODO(nsiow) fix behavior for */? interleaved patterns
func matchWildcard(pattern, value string) bool {
	// Full wildcard case -- '*' matches absolutely everything
	if pattern == "*" {
		return true
	}

	// Empty pattern case -- '' matches absolutely nothing
	if pattern == "" {
		return false
	}

	// Otherwise we do wildcard matches separated by ':' boundaries
	patternSegments := strings.Split(pattern, ":")
	valueSegments := strings.Split(value, ":")

	// Segment length should be the same size
	if len(patternSegments) != len(valueSegments) {
		return false
	}

	// Segments should be equivalent
	for i := range patternSegments {
		p := patternSegments[i]
		v := valueSegments[i]
		if !matchSegment(p, v) {
			return false
		}
	}

	// If we got here, then all our segments matched - success!
	return true
}

// matchSegment handles the comparison of a single segment of an AWS value
func matchSegment(pattern, value string) bool {
	// * matches everything within a subsegment
	if pattern == "*" {
		return true
	}

	// Count the number of '*' and '?'
	wildcards := strings.Count(pattern, "*")
	anys := strings.Count(pattern, "?")

	// If we have both, only choice is a regex
	if wildcards > 0 && anys > 0 {
		return matchViaRegex(pattern, value)
	}

	// If we have neither, treat as string literals
	if wildcards == 0 && anys == 0 {
		return pattern == value
	}

	// Handle wildcards prefixes
	if wildcards == 1 && pattern[0] == '*' {
		return strings.HasSuffix(value, strings.TrimLeft(pattern, "*"))
	}

	// Handle wildcard suffixes
	if wildcards == 1 && pattern[len(pattern)-1] == '*' {
		return strings.HasPrefix(value, strings.TrimRight(pattern, "*"))
	}

	// Handle wildcard prefixes + suffixes
	if wildcards == 2 && pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return strings.Contains(value, strings.Trim(pattern, "*"))
	}

	return matchViaRegex(pattern, value)
}

// matchViaRegex attempts to match the strings via a limited regex subset
func matchViaRegex(pattern, value string) bool {
	pattern = strings.ReplaceAll(pattern, "*", `[^:]*`)
	pattern = strings.ReplaceAll(pattern, "?", `[^:]`)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return re.MatchString(value)
}

// MatchWildcardIgnoreCase determines if the provided string matches the wildcard pattern, using
// AWS's heuristics for wildcards
func matchWildcardIgnoreCase(pattern, value string) bool {
	return matchWildcard(strings.ToLower(pattern), strings.ToLower(value))
}
