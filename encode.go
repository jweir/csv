package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Marshaler is an interface for objects which can Marshal themselves into CSV.
type Marshaler interface {
	MarshalCSV() ([]byte, error)
}

type encoder struct {
	*csv.Writer
	buffer *bytes.Buffer
}

// Marshal returns the CSV encoding of i, which must be a slice of struct types.
//
// Marshal traverses the slice and encodes the primative values.
//
// The first row of the CSV output is a header row. The column names are based
// on the field name.  If a different name is required a struct tag can be used
// to define a new name.
//
//   Field string `csv:"Column Name"`
//
// To skip encoding a field use the "-" as the tag value.
//
//   Field string `csv:"-"`
//
// Boolean fields can use string values to define true or false.
//   Bool bool `true:"Yes" false:"No"`
func Marshal(i interface{}) ([]byte, error) {
	var enc *encoder

	v := reflect.ValueOf(i)

	switch v.Kind() {
	case reflect.Slice:
		el := v.Index(0)
		enc = newEncoder()

		// Write the column headers
		err := enc.Write(colNames(el.Type()))

		if err != nil {
			return []byte{}, err
		}

		n := v.Len()
		for c := 0; c < n; c++ {
			err := enc.Write(encode(v.Index(c)))

			if err != nil {
				return []byte{}, err
			}
		}
	default:
		return []byte{}, errors.New("only slices can be marshalled")
	}

	enc.Flush()
	return enc.buffer.Bytes(), nil
}

func newEncoder() *encoder {
	b := bytes.NewBuffer([]byte{})

	return &encoder{
		buffer: b,
		Writer: csv.NewWriter(b),
	}
}

// colNames takes a struct and returns the computed columns names for each
// field.
func colNames(t reflect.Type) (out []string) {
	l := t.NumField()

	for x := 0; x < l; x++ {
		f := t.Field(x)
		h, ok := fieldHeaderName(f)
		if ok {
			out = append(out, h)
		}
	}

	return
}

// encodes a struct into a CSV row
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
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%+.3g", fv.Complex())
	case reflect.Interface:
		return encodeInterface(fv, st)
	case reflect.Struct:
		return encodeInterface(fv, st)
	default:
		panic(fmt.Sprintf("Unsupported type %s", fv.Kind()))
	}

	return ""
}

func encodeFloat(bits int, f reflect.Value) string {
	return strconv.FormatFloat(f.Float(), 'g', -1, bits)
}

func encodeBool(b bool, st reflect.StructTag) string {
	v := strconv.FormatBool(b)
	tv := st.Get(v)

	if tv != "" {
		return tv
	}
	return v
}

func encodeInterface(fv reflect.Value, st reflect.StructTag) string {
	marshalerType := reflect.TypeOf(new(Marshaler)).Elem()

	if fv.Type().Implements(marshalerType) {
		m := fv.Interface().(Marshaler)
		b, err := m.MarshalCSV()
		if err != nil {
			return ""
		}
		return string(b)
	}

	return ""
}
