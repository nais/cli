package option

type Option[T any] struct {
	isSome bool
	value  T
}

func (o Option[T]) Or(f func() T) Option[T] {
	if o.isSome {
		return o
	}
	return Some(f())
}

func (o Option[T]) OrMaybe(f func() Option[T]) Option[T] {
	if o.isSome {
		return o
	}
	return f()
}

func (o Option[T]) Do(f func(T)) {
	if o.isSome {
		f(o.value)
	}
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func Some[T any](v T) Option[T] {
	return Option[T]{true, v}
}
