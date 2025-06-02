package writer

import (
	"fmt"
	"io"
	"reflect"

	"github.com/pterm/pterm"
)

type tableColumn struct {
	name string
	path string
}

type Table struct {
	o             io.Writer
	columns       []tableColumn
	ignoreHeaders bool
}

func NewTable(o io.Writer) *Table {
	return &Table{
		o: o,
	}
}

func (t *Table) AddColumn(name, path string) {
	t.columns = append(t.columns, tableColumn{name: name, path: path})
}

func (t *Table) SetIgnoreHeaders(ignore bool) {
	t.ignoreHeaders = ignore
}

func (t *Table) Write(v any) error {
	w := pterm.DefaultTable.WithWriter(t.o)

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		panic("expected a slice or array")
	}
	data := make(pterm.TableData, 0, val.Len()+1)

	if !t.ignoreHeaders {
		w = w.WithHasHeader()
		headers := make([]string, len(t.columns))
		for i, col := range t.columns {
			headers[i] = col.name
		}
		data = append(data, headers)
	}

	for i := range val.Len() {
		row := make([]string, len(t.columns))
		for j, col := range t.columns {
			field := val.Index(i).FieldByName(col.path)
			if !field.IsValid() {
				row[j] = ""
				continue
			}
			row[j] = fmt.Sprint(field.Interface())
		}
		data = append(data, row)
	}

	return w.WithData(data).Render()
}
