package output

import (
	"fmt"
	"io"
	"os"
)

type Output interface {
	Println(a ...any)
	Printf(format string, a ...any)
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

// Stdout returns an Output that writes to standard output.
func Stdout() Output {
	return NewWriter(os.Stdout)
}
