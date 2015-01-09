package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Unmarshal(doc []byte, v interface{}) error {
	// x get the headers from the interface
	// x get the headers from the csv
	// x map the interface headers to document headers

	// iterate each row in the doc
	//   create new obj
	//   populate obj with decoded values
	//   append obj to interface

	pv := reflect.ValueOf(v)

	if pv.Kind() != reflect.Ptr || pv.IsNil() {
		return errors.New("type is nil or not a pointer")
	}

	rv := reflect.ValueOf(v).Elem()

	if rv.Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("only slices are allowed: %s", rv.Kind()))
	}

	dec := newDecoder(doc, rv.Type().Elem())

	for {
		row, err := dec.Read()

		if err != nil {
			break
		} else {
			o := reflect.New(dec.Type).Elem()
			dec.set(row, &o)
			rv.Set(reflect.Append(rv, o))
		}

	}

	return nil
}

type decoder struct {
	*csv.Reader
	reflect.Type
	fm   fieldMap
	cols []string
}

func (d *decoder) set(row []string, el *reflect.Value) {
	for i, col := range d.cols {
		val := row[i]
		field := d.fm[col]
		f := el.FieldByName(field.Name)

		switch f.Kind() {
		case reflect.String:
			decodeString(&f, val)
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
			decodeInt(&f, val)
		case reflect.Float32:
			// return decodeFloat(&f, strVal)
		case reflect.Float64:
			// return decodeFloat(64, fv)
		case reflect.Bool:
			// return decodeBool(fv.Bool(), st)
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
			// return fmt.Sprintf("%v", fv.Uint())
		case reflect.Array:
		case reflect.Complex64, reflect.Complex128:
			// return fmt.Sprintf("%+.3g", fv.Complex())
		case reflect.Interface:
			// return decodeInterface(fv, st)
		case reflect.Struct:
			// return decodeInterface(fv, st)
		default:
			panic(fmt.Sprintf("Unsupported type %s", f.Kind()))
		}
	}
}

func decodeInt(f *reflect.Value, val string) error {
	i, e := strconv.Atoi(val)

	if e != nil {
		return e
	}

	f.SetInt(int64(i))
	return nil
}

func decodeString(f *reflect.Value, val string) error {
	f.SetString(val)

	return nil
}

func newDecoder(doc []byte, rt reflect.Type) *decoder {
	b := bytes.NewReader(doc)
	r := csv.NewReader(b)
	ch := colNames(r)
	pf := publicFields(rt)

	return &decoder{
		Reader: r,
		Type:   rt,
		cols:   ch,
		fm:     mapFields(ch, pf),
	}
}

func colNames(c *csv.Reader) []string {
	out, err := c.Read()

	if err != nil {
	}

	return []string(out)
}

type fieldMap map[string]*reflect.StructField

func mapFields(csvHeaders []string, pubFields []*reflect.StructField) fieldMap {
	fm := fieldMap{}

	// seed the fieldMap with accepted columns
	for _, h := range csvHeaders {
		fm[h] = &reflect.StructField{}
	}

	for _, f := range pubFields {
		name, _ := fieldHeaderName(*f)

		if _, ok := fm[name]; ok {
			fm[name] = f
		}
	}

	return fm

}

func publicFields(t reflect.Type) []*reflect.StructField {
	var out []*reflect.StructField

	flen := t.NumField()

	for i := 0; i < flen; i++ {
		sf := t.Field(i)
		if skipField(sf) {
			continue
		}

		out = append(out, &sf)
	}

	return out

}
