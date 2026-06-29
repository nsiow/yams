package wildcard

import (
	"regexp"
	"strings"
	"sync"
)

// regexCache stores compiled regular expressions to avoid recompilation on every match.
// Using sync.Map for concurrent access safety with good performance for read-heavy workloads.
var regexCache sync.Map

// MatchSegments determines if the provided string matches the wildcard pattern, using AWS's
// heuristics for wildcards. Uses index-based iteration to avoid allocations.
func MatchSegments(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return false
	}

	patternSegments := splitSegments(pattern)
	valueSegments := splitSegments(value)
	if len(patternSegments) != len(valueSegments) {
		return false
	}

	for i, pSeg := range patternSegments {
		if !MatchString(pSeg, valueSegments[i]) {
			return false
		}
	}

	return true
}

// MatchSegmentsPreSplit is an optimized version of MatchSegments that accepts pre-split
// value segments to avoid repeated string splitting allocations.
func MatchSegmentsPreSplit(pattern string, valueSegments []string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return false
	}

	patternSegments := splitSegments(pattern)
	if len(patternSegments) != len(valueSegments) {
		return false
	}

	for i, pSeg := range patternSegments {
		if !MatchString(pSeg, valueSegments[i]) {
			return false
		}
	}

	return true
}

func splitSegments(value string) []string {
	if strings.HasPrefix(value, "arn:") {
		return strings.SplitN(value, ":", 6)
	}
	return strings.Split(value, ":")
}

// MatchString handles the comparison of a single segment of an AWS value
func MatchString(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	// Single-pass scan for wildcards and ?s to avoid two strings.Count calls
	wildcards, anys := 0, 0
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			wildcards++
		case '?':
			anys++
		}
	}

	if wildcards > 0 && anys > 0 {
		return matchViaRegex(pattern, value)
	}

	if wildcards == 0 && anys == 0 {
		return pattern == value
	}

	if wildcards == 1 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
	}

	if wildcards == 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}

	if wildcards == 2 && pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	}

	return matchViaRegex(pattern, value)
}

// MatchSegmentsIgnoreCase determines if the provided string matches the wildcard pattern, using
// AWS's heuristics for wildcards. Uses index-based iteration to avoid allocations.
func MatchSegmentsIgnoreCase(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return false
	}

	for {
		pi := strings.IndexByte(pattern, ':')
		vi := strings.IndexByte(value, ':')

		if (pi < 0) != (vi < 0) {
			return false
		}

		var pSeg, vSeg string
		if pi < 0 {
			pSeg, vSeg = pattern, value
		} else {
			pSeg, vSeg = pattern[:pi], value[:vi]
		}

		if !matchStringIgnoreCase(pSeg, vSeg) {
			return false
		}

		if pi < 0 {
			return true
		}
		pattern = pattern[pi+1:]
		value = value[vi+1:]
	}
}

// matchStringIgnoreCase handles the comparison of a single segment with case-insensitive matching
func matchStringIgnoreCase(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	wildcards, anys := 0, 0
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			wildcards++
		case '?':
			anys++
		}
	}

	if wildcards > 0 && anys > 0 {
		return matchViaRegex(strings.ToLower(pattern), strings.ToLower(value))
	}

	if wildcards == 0 && anys == 0 {
		return strings.EqualFold(pattern, value)
	}

	if wildcards == 1 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(value) >= len(suffix) && strings.EqualFold(value[len(value)-len(suffix):], suffix)
	}

	if wildcards == 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix)
	}

	if wildcards == 2 && pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(strings.ToLower(value), strings.ToLower(middle))
	}

	return matchViaRegex(strings.ToLower(pattern), strings.ToLower(value))
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

	// All six ARN components support wildcards per AWS:
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html#Conditions_ARN
	for i := range 5 {
		if !MatchString(patternSegments[i], valueSegments[i]) {
			return false
		}
	}

	patternPath := patternSegments[5]
	valuePath := valueSegments[5]

	return MatchString(patternPath, valuePath)
}

// MatchAllOrNothing performs "all or nothing" wildcard matching
//
// This is defined as allowing wildcards if and only if `pattern = *` (matching everything), but
// no other wildcard matching
func MatchAllOrNothing(pattern, value string) bool {
	return pattern == "*" || pattern == value
}

// matchViaRegex attempts to match the strings via a limited regex subset.
// Compiled regexes are cached to avoid recompilation overhead.
func matchViaRegex(pattern, value string) bool {
	// Check cache first using original pattern as key
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp).MatchString(value)
	}

	// Build regex by escaping literal parts and converting wildcards
	var buf strings.Builder
	buf.WriteString("^")

	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			buf.WriteString(`.*`)
		case '?':
			buf.WriteByte('.')
		default:
			// Find the extent of the literal portion
			j := i
			for j < len(pattern) && pattern[j] != '*' && pattern[j] != '?' {
				j++
			}
			// Escape the literal portion for regex
			buf.WriteString(regexp.QuoteMeta(pattern[i:j]))
			i = j
			continue
		}
		i++
	}

	buf.WriteString("$")

	re, err := regexp.Compile(buf.String())
	if err != nil {
		return false
	}

	regexCache.Store(pattern, re)
	return re.MatchString(value)
}
