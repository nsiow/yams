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

	// Iterate both strings segment by segment without allocating
	for {
		pi := strings.IndexByte(pattern, ':')
		vi := strings.IndexByte(value, ':')

		// Both must have a boundary or both must be at the final segment
		if (pi < 0) != (vi < 0) {
			return false
		}

		var pSeg, vSeg string
		if pi < 0 {
			pSeg, vSeg = pattern, value
		} else {
			pSeg, vSeg = pattern[:pi], value[:vi]
		}

		if !MatchString(pSeg, vSeg) {
			return false
		}

		if pi < 0 {
			return true
		}
		pattern = pattern[pi+1:]
		value = value[vi+1:]
	}
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

	// Iterate pattern segments by index, compare against pre-split value segments
	idx := 0
	for {
		pi := strings.IndexByte(pattern, ':')

		var pSeg string
		if pi < 0 {
			pSeg = pattern
		} else {
			pSeg = pattern[:pi]
		}

		if idx >= len(valueSegments) {
			return false
		}
		if !MatchString(pSeg, valueSegments[idx]) {
			return false
		}
		idx++

		if pi < 0 {
			return idx == len(valueSegments)
		}
		pattern = pattern[pi+1:]
	}
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
			buf.WriteString(`[^:]*`)
		case '?':
			buf.WriteString(`[^:]`)
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
