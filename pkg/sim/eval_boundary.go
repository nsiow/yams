package sim

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalPermissionsBoundary assesses the permissions boundary of the Principal to determine whether
// or not it allows the provided AuthContext
func evalPermissionsBoundary(trc *trace.Trace, opt *Options, ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating permission boundaries")
	defer trc.Pop()

	// Empty permissions boundary = allowed; otherwise we have to evaluate
	if ac.Principal.PermissionsBoundary.Empty() {
		effectset := EffectSet{}
		effectset.Add(policy.EFFECT_ALLOW)
		return effectset, nil
	}

	// Specify the statement evaluation funcs we will consider for permission boundary access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesResource,
		evalStatementMatchesCondition,
	}

	return evalPolicy(trc, opt, ac, ac.Principal.PermissionsBoundary, funcs)
}
