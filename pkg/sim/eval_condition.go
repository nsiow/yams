package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// evalCheckCondition assesses a single condition operator to determine whether or not it matches
// the provided AuthContext
func evalCheckCondition(s *subject, op string, cond map[string]policy.Value) (bool, error) {

	// TODO(nsiow) implement PushWithAttr so that `op` is in a more appropriate context?
	s.trc.Push("evaluating Operation")
	s.trc.Attr("op", op)
	defer s.trc.Pop()

	// An empty condition should actually evaluate to false
	if len(cond) == 0 {
		s.trc.Observation("no match; empty condition")
		return false, nil
	}

	// Check to see if the condition operator is supported
	f, exists := ResolveConditionEvaluator(op)
	if !exists {
		if s.opts.SkipUnknownConditionOperators {
			return true, nil
		} else {
			s.trc.Observation("no match; unknown condition")
			return false, fmt.Errorf("unknown condition operator '%s'", op)
		}
	}

	// Check condition evaluation against actual values
	for k, v := range cond {
		match := f(s, k, v)
		if !match {
			s.trc.Observation("no match; condition evaluated to false")
			return false, nil
		}
	}

	s.trc.Observation("match!")
	return true, nil
}
