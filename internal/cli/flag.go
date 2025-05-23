package cli

type count int

type flagTypes interface{ int | bool | string | count }

type Flag[T flagTypes] struct {
	name   string
	usage  string
	short  string
	sticky bool
	value  *T
}

func NewFlag[T flagTypes](name, usage, short string, value *T) Flag[T] {
	return Flag[T]{
		name:   name,
		usage:  usage,
		value:  value,
		short:  short,
		sticky: false,
	}
}

func NewStickyFlag[T flagTypes](name, usage, short string, value *T) Flag[T] {
	return Flag[T]{
		name:   name,
		usage:  usage,
		value:  value,
		short:  short,
		sticky: true,
	}
}
