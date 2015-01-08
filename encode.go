package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
)

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
	case reflect.Array:
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
