package wildcard

import (
	"regexp"
	"strings"
)

// MatchSegments determines if the provided string matches the wildcard pattern, using AWS's
// heuristics for wildcards
//
// TODO(nsiow) consider moving this to its own package
// TODO(nsiow) add trace logging for better debugging
// TODO(nsiow) fix behavior for */? interleaved patterns
func MatchSegments(pattern, value string) bool {
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
		if !MatchString(p, v) {
			return false
		}
	}

	// If we got here, then all our segments matched - success!
	return true
}

// MatchString handles the comparison of a single segment of an AWS value
func MatchString(pattern, value string) bool {
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

	// Fall back to regex matching
	return matchViaRegex(pattern, value)
}

// MatchSegmentsIgnoreCase determines if the provided string matches the wildcard pattern, using
// AWS's heuristics for wildcards
func MatchSegmentsIgnoreCase(pattern, value string) bool {
	return MatchSegments(strings.ToLower(pattern), strings.ToLower(value))
}

// MatchArn performs specialized ARN-matching logic for certain condition operators
func MatchArn(pattern, value string) bool {
	// TODO(nsiow) confirm that "*" actually matches all Principals... I am not sure of this
	if pattern == "*" {
		return true
	}

	// TODO(nsiow) check the value of 6
	// arn:aws:iam:us-east-1:account:role/foo
	patternSegments := strings.SplitN(pattern, ":", 6)
	valueSegments := strings.SplitN(value, ":", 6)

	// Segment length should be valid
	if len(patternSegments) != 6 || len(valueSegments) != 6 {
		return false
	}

	// Leader should be the same
	if patternSegments[0] != valueSegments[0] {
		return false
	}

	// Partition should be the same
	if patternSegments[1] != valueSegments[1] {
		return false
	}

	// Service should be the same
	if patternSegments[2] != valueSegments[2] {
		return false
	}

	// Region can be wildcarded
	if !MatchString(patternSegments[3], valueSegments[3]) {
		return false
	}

	// Account should be the same
	if patternSegments[4] != valueSegments[4] {
		return false
	}

	patternPath := patternSegments[5]
	valuePath := valueSegments[5]

	// Resource type should be the same
	patternType, newPattern, patternFound := strings.Cut(patternPath, "/")
	valueType, newValue, valueFound := strings.Cut(valuePath, "/")
	switch {
	case patternFound && valueFound:
		if patternType != valueType {
			return false
		}
		patternPath = newPattern
		valuePath = newValue
	case !patternFound && !valueFound:
		break
	case patternFound != valueFound:
		return false
	}

	return MatchString(patternPath, valuePath)
}

// MatchAllOrNothing performs "all or nothing" wildcard matching
//
// This is defined as allowing wildcards if and only if `pattern = *` (matching everything), but
// no other wildcard matching
func MatchAllOrNothing(pattern, value string) bool {
	return pattern == "*" || pattern == value
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
