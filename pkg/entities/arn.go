package entities

import "strings"

const (
	arnSegmentCount     = 6
	arnSegmentSeparator = ":"
)

func arnComponents(arn string) []string {
	return strings.SplitN(arn, arnSegmentSeparator, arnSegmentCount)
}

func arnComponent(arn string, idx int) string {
	components := arnComponents(arn)
	if len(components) <= idx {
		return ""
	}

	return components[idx]
}

func arnPartition(arn string) string {
	return arnComponent(arn, 1)
}

func arnService(arn string) string {
	return arnComponent(arn, 2)
}

func arnRegion(arn string) string {
	return arnComponent(arn, 3)
}

func arnAccountId(arn string) string {
	return arnComponent(arn, 4)
}

func arnResourceSegment(arn string) string {
	return arnComponent(arn, 5)
}

// TODO(nsiow) this is very incomplete
func arnResourceType(arn string) string {
	seg := arnResourceSegment(arn)

	switch {
	case arnService(arn) == "s3" && strings.Contains(arnResourceId(arn), "/"):
		return "object"
	case arnService(arn) == "s3" && !strings.Contains(arnResourceId(arn), "/"):
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
func arnResourceId(arn string) string {
	seg := arnResourceSegment(arn)

	switch {
	case arnService(arn) == "s3" && len(arnRegion(arn)) == 0:
		return seg
	case strings.Contains(seg, "/"):
		return strings.SplitN(seg, "/", 2)[1]
	case strings.Contains(seg, ":"):
		return strings.SplitN(seg, ":", 2)[1]
	default:
		return seg
	}
}
