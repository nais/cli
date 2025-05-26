package output

import (
	"fmt"
	"io"
)

type Output interface {
	Println(...any)
	Printf(string, ...any)
}

type writer struct {
	w io.Writer
}

func (w *writer) Println(a ...any) {
	_, _ = fmt.Fprintln(w.w, a...)
}

func (w *writer) Printf(format string, a ...any) {
	_, _ = fmt.Fprintf(w.w, format, a...)
}

func NewWriter(w io.Writer) Output {
	return &writer{w: w}
}
