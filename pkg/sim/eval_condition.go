package sim

import (
	"github.com/nsiow/yams/pkg/policy"
)

// evalCheckCondition assesses a single condition operator to determine whether or not it matches
// the provided AuthContext
func evalCheckCondition(s *subject, op string, cond policy.ConditionValues) bool {

	s.trc.Push("evaluating operation: %s", op)
	defer s.trc.Pop()

	// An empty condition should actually evaluate to false
	if len(cond) == 0 {
		s.trc.Log("no match; empty condition")
		return false
	}

	// Check to see if the condition operator is supported
	f, exists := ResolveConditionEvaluator(op)
	if !exists {
		s.trc.Log("no match; unknown condition operator: %s", op)
		return false
	}

	// Check condition evaluation against actual values
	for k, v := range cond {
		match := f(s, k, v)
		if !match {
			s.trc.Log("no match; condition evaluated to false")
			return false
		}
	}

	s.trc.Log("all condition pairs matched for op: %s", op)
	return true
}
