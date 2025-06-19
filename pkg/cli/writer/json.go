package writer

import (
	"encoding/json"
	"io"
)

type JSON struct {
	prettify bool
	o        io.Writer
}

func NewJSON(o io.Writer, prettify bool) *JSON {
	return &JSON{
		prettify: prettify,
		o:        o,
	}
}

func (j *JSON) Write(v any) error {
	enc := json.NewEncoder(j.o)
	if j.prettify {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}
