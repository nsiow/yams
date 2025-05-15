package arn

import "strings"

const (
	arnSegmentCount     = 6
	arnSegmentSeparator = ":"
)

func Components(arn string) []string {
	return strings.SplitN(arn, arnSegmentSeparator, arnSegmentCount)
}

func Component(arn string, idx int) string {
	components := Components(arn)
	if len(components) <= idx {
		return ""
	}

	return components[idx]
}

func Partition(arn string) string {
	return Component(arn, 1)
}

func Service(arn string) string {
	return Component(arn, 2)
}

func Region(arn string) string {
	return Component(arn, 3)
}

func Account(arn string) string {
	return Component(arn, 4)
}

func ResourceSegment(arn string) string {
	return Component(arn, 5)
}

// TODO(nsiow) this is very incomplete
func ResourcePath(arn string) string {
	seg := ResourceSegment(arn)

	switch {
	case Service(arn) == "s3" && strings.Contains(ResourceId(arn), "/"):
		return "object"
	case Service(arn) == "s3" && !strings.Contains(ResourceId(arn), "/"):
		return "bucket"
	case strings.Contains(seg, ":"):
		return strings.SplitN(seg, ":", 2)[0]
	case strings.Contains(seg, "/"):
		return strings.SplitN(seg, "/", 2)[0]
	default:
		return ""
	}
}

// TODO(nsiow) this is very incomplete
func ResourceId(arn string) string {
	seg := ResourceSegment(arn)

	switch {
	case Service(arn) == "s3" && len(Region(arn)) == 0:
		return seg
	case strings.Contains(seg, "/"):
		return strings.SplitN(seg, "/", 2)[1]
	case strings.Contains(seg, ":"):
		return strings.SplitN(seg, ":", 2)[1]
	default:
		return seg
	}
}
