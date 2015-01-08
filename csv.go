// Package csv provides `Marshal` and `UnMarshal` encoding functions for CSV(Comma Seperated Value) data.
// This package is built on the the standard library's encoding/csv.
package csv

import (
	"errors"
	"reflect"
)

type Marshaler interface {
	MarshalCSV() ([]byte, error)
}

// Marshal returns the CSV encoding of i, which must be a slice struct types.
//
// Marshal traverses the slice and encodes the primative values.
//
// The first row of the CSV output is a header row. The column names are based
// on the field name.  If a different name is required a struct tag can be used to define a new name.
//   Field string `csv:"Column Name"`
//
// To skip encoding a field use the "-" as the tag value.
//   Field string `csv:"-"`
//
func Marshal(i interface{}) ([]byte, error) {
	enc := newEncoder()

	v := reflect.ValueOf(i)

	switch v.Kind() {
	case reflect.Slice:
		e := v.Index(0)
		enc.Write(headers(e.Type()))

		n := v.Len()
		for x := 0; x < n; x++ {
			enc.Write(encode(v.Index(x)))
		}
	default:
		return []byte{}, errors.New("Only slices can be marshalled")
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

func skipField(f reflect.StructField) bool {
	if f.Tag.Get("csv") == "-" {
		return true
	}

	return false
}

func header(f reflect.StructField) (string, bool) {
	h := f.Tag.Get("csv")

	if h == "-" {
		return "", false
	}

	// If there is no tag set, use a default name
	if h == "" {
		return f.Name, true
	}

	return h, true
}

func headers(t reflect.Type) (out []string) {
	l := t.NumField()

	for x := 0; x < l; x++ {
		f := t.Field(x)
		h, ok := header(f)
		if ok {
			out = append(out, h)
		}
	}

	return
}
