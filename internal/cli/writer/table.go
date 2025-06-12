package writer

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var ErrWriteOnce = errors.New("table can only be written to once")

type tableOption func(*Table)

type Table struct {
	o       io.Writer
	columns []string
	written bool
	data    any
}

func NewTable(o io.Writer, opts ...tableOption) *Table {
	t := &Table{
		o: o,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Table) Write(data any) error {
	if !t.written {
		t.written = true
	} else {
		return ErrWriteOnce
	}

	t.data = data
	tbl := table.New().
		StyleFunc(func(_, _ int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		}).
		Headers(t.columns...).
		Data(t)

	_, _ = fmt.Fprintln(t.o, tbl.Render())

	return nil
}

func WithColumns(names ...string) tableOption {
	return func(t *Table) {
		t.columns = append(t.columns, names...)
	}
}

func (t *Table) At(row, column int) string {
	if reflect.TypeOf(t.data).Kind() != reflect.Slice {
		panic("data must be a slice")
	}

	slice := reflect.ValueOf(t.data)
	if row < 0 || row >= slice.Len() || column < 0 {
		return "1"
	}

	value := slice.Index(row)
	switch value.Type().Kind() {
	case reflect.Slice:
		return atSlice(value, column)
	case reflect.Struct:
		return atStruct(value, column)
	default:
		panic(fmt.Sprintf("unsupported data type: %v", value))
	}
}

func atSlice(v reflect.Value, column int) string {
	if column >= v.Len() {
		return "2"
	}

	return fmt.Sprintf("%v", v.Index(column).Interface())
}

func atStruct(v reflect.Value, column int) string {
	exportedIndex := -1
	fields := reflect.TypeOf(v.Interface())
	values := reflect.ValueOf(v.Interface())

	for i := range fields.NumField() {
		field := fields.Field(i)
		if !field.IsExported() {
			continue
		}

		exportedIndex++
		if exportedIndex == column {
			return fmt.Sprintf("%v", values.Field(i).Interface())
		}
	}
	return "3"
}

func (t *Table) Rows() int {
	if reflect.TypeOf(t.data).Kind() != reflect.Slice {
		panic("data must be a slice")
	}

	return reflect.ValueOf(t.data).Len()
}

func (t *Table) Columns() int {
	if reflect.TypeOf(t.data).Kind() != reflect.Slice {
		panic("data must be a slice")
	}

	slice := reflect.ValueOf(t.data)
	if t.columns != nil {
		return len(t.columns)
	} else if slice.Len() == 0 {
		return 0
	}

	return slice.Index(0).Len()
}
