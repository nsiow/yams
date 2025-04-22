package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalCheckCondition assesses a single condition operator to determine whether or not it matches
// the provided AuthContext
func evalCheckCondition(s *subject, op string, cond policy.ConditionValues) bool {

	// TODO(nsiow) implement PushWithAttr so that `op` is in a more appropriate context?
	s.trc.Push("evaluating Operation")
	defer s.trc.Pop()

	// An empty condition should actually evaluate to false
	if len(cond) == 0 {
		s.trc.Observation("no match; empty condition")
		return false
	}

	// Check to see if the condition operator is supported
	f, exists := ResolveConditionEvaluator(op)
	if !exists {
		s.trc.Observation("no match; unknown condition operator: ")
		return false
	}

	// Check condition evaluation against actual values
	for k, v := range cond {
		match := f(s, k, v)
		if !match {
			s.trc.Observation("no match; condition evaluated to false")
			return false
		}
	}

	s.trc.Observation("match!")
	return true
}
