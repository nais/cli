package output

import (
	"fmt"
	"io"
)

type Output interface {
	Println(...any)
}

type writer struct {
	w io.Writer
}

func (w *writer) Println(a ...any) {
	_, _ = fmt.Fprintln(w.w, a...)
}

func NewWriter(w io.Writer) Output {
	return &writer{w: w}
}
