package sim

import (
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalCheckCondition assesses a single condition operator to determine whether or not it matches
// the provided AuthContext
func evalCheckCondition(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext,
	op string,
	cond map[string]policy.Value) (bool, error) {

	// TODO(nsiow) implement PushWithAttr so that `op` is in a more appropriate context?
	trc.Push("evaluating Operation")
	trc.Attr("op", op)
	defer trc.Pop()

	// An empty condition should actually evaluate to false
	if len(cond) == 0 {
		trc.Observation("no match; empty condition")
		return false, nil
	}

	// Check to see if the condition operator is supported
	f, exists := ResolveConditionEvaluator(op)
	if !exists {
		if opt.FailOnUnknownCondition {
			trc.Observation("no match; unknown condition")
			return false, fmt.Errorf("unknown condition operator '%s'", op)
		}
		return true, nil
	}

	// Check condition evaluation against actual values
	for k, v := range cond {
		match := f(ac, trc, k, v)
		if !match {
			trc.Observation("no match; condition evaluated to false")
			return false, nil
		}
	}

	trc.Observation("match!")
	return true, nil
}
