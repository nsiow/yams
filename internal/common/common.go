package common

// Map is a generic slice-mapping function performing the operation f(S[I], f(I)O) -> S[O]
func Map[I, O any](in []I, f func(I) O) []O {
	out := make([]O, len(in))
	for i, e := range in {
		out[i] = f(e)
	}

	return out
}

// MapWithErr provides the same functionality as [Map] but with optional error handling
func MapWithErr[I, O any](in []I, f func(I) (O, error)) ([]O, error) {
	var err error

	out := make([]O, len(in))
	for i, e := range in {
		out[i], err = f(e)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
