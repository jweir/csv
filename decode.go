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

type decoder struct {
	*csv.Reader                // the csv document for input
	reflect.Type               // the underlying struct to decode
	out          reflect.Value // the slice output
	fms          []fieldColMap //
	cols         []string      // colum names
}

type decoderFn func(*reflect.Value, string) error

// maps a CSV column Name and index to a StructField
type fieldColMap struct {
	colName     string
	colIndex    int
	structField *reflect.StructField
	decode      decoderFn
}

// UnMarshaller is the interface implemented by objects which can unmarshall the CSV row itself.
type Unmarshaler interface {
	UnmarshalCSV([]string, []string) error
}

// Unmarshal decodes the CSV document into the slice interface.
func Unmarshal(doc []byte, v interface{}) error {
	if err := checkValidInterface(v); err != nil {
		return err
	}

	rv := reflect.ValueOf(v).Elem()
	dec, err := newDecoder(doc, rv)

	if err != nil {
		return err
	}

	dec.unmarshal()
	return nil
}

func (dec *decoder) unmarshal() error {
	for {
		row, err := dec.Read()

		if err != nil {
			break
		} else {
			o := reflect.New(dec.Type).Elem()
			err := dec.set(row, &o)
			if err != nil {
				return err
			}
			dec.out.Set(reflect.Append(dec.out, o))
		}

	}

	return nil

}

func checkValidInterface(v interface{}) error {
	pv := reflect.ValueOf(v)

	if pv.Kind() != reflect.Ptr || pv.IsNil() {
		return errors.New("type is nil or not a pointer")
	}

	rv := reflect.ValueOf(v).Elem()

	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("only slices are allowed: %s", rv.Kind())
	}

	return nil
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

			fm.assignDecoder()

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

func newDecoder(doc []byte, rv reflect.Value) (*decoder, error) {
	b := bytes.NewReader(doc)
	r := csv.NewReader(b)
	cols, err := r.Read()

	if err != nil {
		return nil, err
	}

	ch := []string(cols)
	el := rv.Type().Elem()

	return &decoder{
		Reader: r,
		Type:   el,
		out:    rv,
		fms:    mapFieldsToCols(el, ch),
	}, nil
}

func assign(fm *fieldColMap, fn decoderFn) {
	fm.decode = func(f *reflect.Value, v string) error {
		return fn(f, v)
	}
}

func (fm *fieldColMap) assignDecoder() {
	switch fm.structField.Type.Kind() {
	case reflect.String:
		assign(fm, fm.decodeString)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		assign(fm, fm.decodeInt)
	case reflect.Float32:
		assign(fm, fm.decodeFloat(32))
	case reflect.Float64:
		assign(fm, fm.decodeFloat(64))
	case reflect.Bool:
		assign(fm, fm.decodeBool)
	default:
		assign(fm, fm.ignoreValue)
	}
}

func (dec *decoder) set(row []string, el *reflect.Value) error {
	for _, fm := range dec.fms {
		val := row[fm.colIndex]
		field := fm.structField

		f := el.FieldByName(field.Name)
		if fm.decode != nil {
			err := fm.decode(&f, val)

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no decoder for %s\n", val)
		}
	}

	return nil
}

func (fm *fieldColMap) decodeBool(f *reflect.Value, val string) error {
	var bv bool

	bt := fm.structField.Tag.Get("true")
	bf := fm.structField.Tag.Get("false")

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

func (fm *fieldColMap) decodeInt(f *reflect.Value, val string) error {
	i, e := strconv.Atoi(val)

	if e != nil {
		return e
	}

	f.SetInt(int64(i))
	return nil
}

func (fm *fieldColMap) decodeString(f *reflect.Value, val string) error {
	f.SetString(val)

	return nil
}

func (fm *fieldColMap) decodeFloat(bit int) decoderFn {
	return func(f *reflect.Value, val string) error {
		n, err := strconv.ParseFloat(val, bit)

		if err != nil {
			return err
		}

		f.SetFloat(n)

		return nil
	}
}

// ignoreValue does nothing. This is for unsupported types.
func (fm *fieldColMap) ignoreValue(f *reflect.Value, val string) error {
	return nil
}
