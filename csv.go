package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
)

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
	var o string
	l := v.Type().NumField()

	for x := 0; x < l; x++ {
		f := v.Field(x)

		switch f.Kind() {
		case reflect.String:
			o = f.String()
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
			o = fmt.Sprintf("%v", f.Int())
		case reflect.Float32:
			o = encodeFloat(32, f)
		case reflect.Float64:
			o = encodeFloat(64, f)
		default:
			o = ""
		}

		out = append(out, o)
	}

	return
}

func encodeFloat(bits int, f reflect.Value) string {
	return strconv.FormatFloat(f.Float(), 'g', -1, bits)
}

func header(t reflect.Type, field string) (string, error) {
	// Ignore if the field exists, it should since only the fields
	// in the interface are being read
	f, _ := t.FieldByName(field)

	h := f.Tag.Get("csv")

	// If there is no tag set, use a default name
	if h == "" {
		return field, nil
	}

	return h, nil
}

func headers(t reflect.Type) (out []string) {
	l := t.NumField()

	for x := 0; x < l; x++ {
		f := t.Field(x)
		h, _ := header(t, f.Name)
		out = append(out, h)
	}

	return
}
