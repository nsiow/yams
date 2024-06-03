package sim

// AuthContext defines the tertiary context of a request that can be used for authz decisions
// TODO(nsiow) figure out if auth context can be expanded so that it handles the majority of
// ConditionKeys... also if that happens do we have to pass p, r, s etc to every single function?
type AuthContext struct {
	Properties map[string][]string
}
