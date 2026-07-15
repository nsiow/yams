package v1

func (api *API) requestContext(input map[string]string, disableSharedContext bool) map[string]string {
	if disableSharedContext || len(api.SharedContext) == 0 {
		return input
	}

	ctx := make(map[string]string, len(api.SharedContext)+len(input))
	for k, v := range api.SharedContext {
		ctx[k] = v
	}
	for k, v := range input {
		ctx[k] = v
	}
	return ctx
}
