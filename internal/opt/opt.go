package opt

type Option[T any] struct {
	v   T
	has bool
}

func Some[T any](v T) Option[T] {
	return Option[T]{v: v, has: true}
}

func None[T any]() Option[T] {
	return Option[T]{has: false}
}

func (o *Option[T]) Get() (T, bool) {
	return o.v, o.has
}

func (o *Option[T]) GetOrDefault(def T) T {
	if o.has {
		return o.v
	}

	return def
}
