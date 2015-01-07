// Package csv provides `Marshal` and `UnMarshal` encoding functions for CSV(Comma Seperated Value) data.
// This package is built on the the standard library's encoding/csv.
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
)

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
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

type encoder struct {
	*csv.Writer
	buffer *bytes.Buffer
}

func newEncoder() encoder {
	b := bytes.NewBuffer([]byte{})

	return encoder{
		buffer: b,
		Writer: csv.NewWriter(b),
	}
}

func encode(v reflect.Value) (out []string) {
	l := v.Type().NumField()

	for x := 0; x < l; x++ {
		fv := v.Field(x)
		st := v.Type().Field(x).Tag

		if st.Get("csv") == "-" {
			continue
		}
		o := encodeFieldValue(fv, st)
		out = append(out, o)
	}

	return
}

// Returns the string representation of the field value
func encodeFieldValue(fv reflect.Value, st reflect.StructTag) string {
	switch fv.Kind() {
	case reflect.String:
		return fv.String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return fmt.Sprintf("%v", fv.Int())
	case reflect.Float32:
		return encodeFloat(32, fv)
	case reflect.Float64:
		return encodeFloat(64, fv)
	case reflect.Bool:
		return encodeBool(fv.Bool(), st)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return fmt.Sprintf("%v", fv.Uint())
	case reflect.Array:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Interface:
	// time.Time
	default:
		panic(fmt.Sprintf("Unsupported type %s", fv.Kind()))
	}

	return ""
}

func encodeFloat(bits int, f reflect.Value) string {
	return strconv.FormatFloat(f.Float(), 'g', -1, bits)
}

func skipField(f reflect.StructField) bool {
	if f.Tag.Get("csv") == "-" {
		return true
	}

	return false
}

func encodeBool(b bool, st reflect.StructTag) string {
	v := strconv.FormatBool(b)
	tv := st.Get(v)

	if tv != "" {
		return tv
	}
	return v
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
