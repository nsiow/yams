package entities

import "strings"

const (
	arnSegmentCount     = 6
	arnSegmentSeparator = ":"
)

// Arn represents an AWS Resource Name, a unique identifier for a cloud resource in AWS
type Arn string

func (a Arn) components() []string {
	return strings.SplitN(a.String(), arnSegmentSeparator, arnSegmentCount)
}

func (a Arn) component(idx int) string {
	components := a.components()
	if len(components) <= idx {
		return ""
	}

	return components[idx]
}

func (a Arn) String() string {
	return string(a)
}

func (a *Arn) Partition() string {
	return a.component(1)
}

func (a *Arn) Service() string {
	return a.component(2)
}

func (a *Arn) Region() string {
	return a.component(3)
}

func (a *Arn) AccountId() string {
	return a.component(4)
}

func (a *Arn) ResourceSegment() string {
	return a.component(5)
}

// TODO(nsiow) this is very incomplete
func (a *Arn) ResourceType() string {
	seg := a.ResourceSegment()

	switch {
	case a.Service() == "s3" && len(a.Region()) == 0 && strings.Contains(a.ResourceId(), "/"):
		return "object"
	case a.Service() == "s3" && len(a.Region()) == 0 && !strings.Contains(a.ResourceId(), "/"):
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
func (a *Arn) ResourceId() string {
	seg := a.ResourceSegment()
	switch {
	case a.Service() == "s3" && len(a.Region()) == 0:
		return seg
	case strings.Contains(seg, "/"):
		return strings.SplitN(seg, "/", 2)[1]
	case strings.Contains(seg, ":"):
		return strings.SplitN(seg, ":", 2)[1]
	default:
		return seg
	}
}
