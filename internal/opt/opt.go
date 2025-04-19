package opt

type Option[T any] struct {
	v        T
	hasValue bool
}

func (o *Option[T]) Get() (T, bool) {
	return o.v, o.hasValue
}

func (o *Option[T]) GetOrDefault(def T) T {
	if o.hasValue {
		return o.v
	}

	return def
}

func Some[T any](v T) Option[T] {
	return Option[T]{v: v, hasValue: true}
}

func None[T any]() Option[T] {
	return Option[T]{hasValue: false}
}
