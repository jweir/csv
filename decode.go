package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(doc []byte, v interface{}) error {
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

// maps a CSV column Name and index to a StructField
type fieldColMap struct {
	colName     string
	colIndex    int
	structField *reflect.StructField
}

// colNames reuturns the CSV header column names
func colNames(c *csv.Reader) []string {
	out, err := c.Read()

	if err != nil {
	}

	return []string(out)
}

// mapFields creates a set of fieldMap instrances where
// the CSV colnames and the exported field names intersect
func mapFieldsToCols(t reflect.Type, cols []string) []fieldColMap {
	pFields := exportedFields(t)

	cMap := map[string]int{}
	fMap := []fieldColMap{}

	for i, col := range cols {
		cMap[col] = i
	}

	for _, f := range pFields {
		name, ok := fieldHeaderName(*f)

		if ok == false {
			continue
		}

		index, ok := cMap[name]

		if ok == true {
			fm := fieldColMap{
				colName:     name,
				colIndex:    index,
				structField: f,
			}

			fMap = append(fMap, fm)
		}
	}

	return fMap
}

func exportedFields(t reflect.Type) []*reflect.StructField {
	var out []*reflect.StructField

	v := reflect.New(t).Elem()
	flen := v.NumField()

	for i := 0; i < flen; i++ {

		sf := t.Field(i)

		if skipField(sf) {
			continue
		}

		// Work around issue with CanSet not working on struct fields
		c := string(sf.Name[0])
		if c == strings.ToUpper(c) {
			out = append(out, &sf)
		}
	}

	return out

}

type decoder struct {
	*csv.Reader
	reflect.Type
	fms  []fieldColMap
	cols []string
}

func newDecoder(doc []byte, rt reflect.Type) *decoder {
	b := bytes.NewReader(doc)
	r := csv.NewReader(b)
	ch := colNames(r)

	return &decoder{
		Reader: r,
		Type:   rt,
		fms:    mapFieldsToCols(rt, ch),
	}
}

func (d *decoder) set(row []string, el *reflect.Value) {
	for _, fm := range d.fms {
		val := row[fm.colIndex]
		field := fm.structField

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
			decodeBool(&f, val, field.Tag)
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

func decodeBool(f *reflect.Value, val string, tag reflect.StructTag) error {
	var bv bool

	bt := tag.Get("true")
	bf := tag.Get("false")

	switch val {
	case bt:
		bv = true
	case bf:
		bv = false
	default:
		bv = true
	}

	f.SetBool(bv)

	return nil
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
