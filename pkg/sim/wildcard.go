package sim

import (
	"regexp"
	"strings"
)

// MatchWildcard determines if the provided string matches the wildcard pattern, using AWS's
// heuristics for wildcards
//
// TODO(nsiow) add trace logging for better debugging
func matchWildcard(pattern, value string) bool {
	// Full wildcard case -- '*' matches absolutely everything
	if pattern == "*" {
		return true
	}

	// Otherwise we do wildcard matches separated by ':' boundaries
	patternSegments := strings.Split(pattern, ":")
	valueSegments := strings.Split(value, ":")

	// Segment length should be the same size
	if len(patternSegments) != len(valueSegments) {
		return false
	}

	for i := range patternSegments {
		p := patternSegments[i]
		v := valueSegments[i]

		// * matches everything within a subsegment
		if p == "*" {
			continue
		}

		// Count the number of wildcards
		wildcards := strings.Count(p, "*")

		// Handle no wildcards, string literals
		if wildcards == 0 {
			if p != v {
				return false
			}
		}

		// Handle wildcards prefixes
		if wildcards == 1 && p[0] == '*' {
			if !strings.HasSuffix(v, strings.TrimLeft(p, "*")) {
				return false
			}
		}

		// Handle wildcard suffixes
		if wildcards == 1 && p[len(p)-1] == '*' {
			if !strings.HasPrefix(v, strings.TrimRight(p, "*")) {
				return false
			}
		}

		// Handle wildcard prefixes + suffixes
		if wildcards == 2 && p[0] == '*' && p[len(p)-1] == '*' {
			if !strings.Contains(v, strings.Trim(p, "*")) {
				return false
			}
		}

		// Otherwise, defer to regex matching; ouch!
		re, err := convertWildcardToRegex(p)
		if err != nil {
			return false
		}
		if !re.MatchString(v) {
			return false
		}
	}

	// If we got here, then all our segments matched - success!
	return true
}

// convertWildcardToRegex takes an AWS IAM wildcard pattern and converts it into a best-effort
// regexp
func convertWildcardToRegex(pattern string) (*regexp.Regexp, error) {
	pattern = strings.ReplaceAll(pattern, "*", `[^:]*`)
	pattern = strings.ReplaceAll(pattern, "?", `[^:]?`)
	return regexp.Compile(pattern)
}

// MatchWildcardIgnoreCase determines if the provided string matches the wildcard pattern, using
// AWS's heuristics for wildcards
func matchWildcardIgnoreCase(pattern, value string) bool {
	return matchWildcard(strings.ToLower(pattern), strings.ToLower(value))
}
