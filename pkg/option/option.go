package option

import "fmt"

type Option[T any] struct {
	isSome bool
	value  T
}

func (o Option[T]) OrValue(v T) Option[T] {
	if o.isSome {
		return o
	}
	return Some(v)
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

func (o Option[T]) String() string {
	if o.isSome {
		return fmt.Sprintf("%v", o.value)
	}
	return ""
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func Some[T any](v T) Option[T] {
	return Option[T]{true, v}
}
