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

// Row is one row of CSV data, indexed by column name or position
type Row struct {
	Columns *[]string // The name of the columns, in order
	Data    []string  // the data for the row
}

// At returns the rows data for the column positon i
func (r *Row) At(i int) string {
	return r.Data[i]
}

// Named reutrns the row's data for the first columne named 'n'
func (r *Row) Named(n string) (string, error) {
	for i, cn := range *r.Columns {
		if cn == n {
			return r.At(i), nil
		}
	}

	return "", fmt.Errorf("No column found for %s", n)
}

type decoder struct {
	*csv.Reader                // the csv document for input
	reflect.Type               // the underlying struct to decode
	out          reflect.Value // the slice output
	fms          []cfield      //
	cols         []string      // colum names
}

type decoderFn func(*reflect.Value, *Row) error

// maps a CSV column Name and index to a StructField
type cfield struct {
	colName     string
	colIndex    int
	structField *reflect.StructField
	decode      decoderFn
}

// Unmarshaler is the interface implemented by objects which can unmarshall the CSV row itself.
type Unmarshaler interface {

	// UnmarshalCSV receives a string with the column value matching this field
	// and a reference to the the current row.
	// This allows composing a value from mutliple columns.
	UnmarshalCSV(string, *Row) error
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
		raw, err := dec.Read()

		if err != nil {
			break
		} else {
			row := dec.newRow(raw)
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

func (dec *decoder) newRow(raw []string) *Row {
	return &Row{
		Columns: &dec.cols,
		Data:    raw,
	}
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

// interface is implemented on a value
const impsVal int = 1

// interface is implemented on a pointer
const impsPtr int = 2

// checks if an object implements the Unmarshaler interface
func impsUnmarshaller(et reflect.Type, i interface{}) (int, error) {
	el := reflect.New(et).Elem()
	it := reflect.TypeOf(i).Elem()

	if el.Type().Implements(it) {
		return impsVal, nil
	}

	if el.Addr().Type().Implements(it) {
		return impsPtr, nil
	}

	return 0, fmt.Errorf("%v el does not implement %s", el, it.Name())
}

// mapFields creates a set of fieldMap instrances where
// the CSV colnames and the exported field names intersect
func (dec *decoder) mapFieldsToCols(t reflect.Type, cols []string) {
	pFields := exportedFields(t)

	cMap := map[string]int{}

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
			fm := cfield{
				colName:     name,
				colIndex:    index,
				structField: f,
			}

			if code, err := impsUnmarshaller(f.Type, new(Unmarshaler)); err == nil {
				fm.assignUnmarshaller(code)
			} else {
				fm.assignDecoder()
			}

			dec.fms = append(dec.fms, fm)
		}
	}

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

	dec := decoder{
		Reader: r,
		Type:   el,
		out:    rv,
		cols:   cols,
	}

	dec.mapFieldsToCols(el, ch)

	return &dec, nil
}

func assign(fm *cfield, fn decoderFn) {
	fm.decode = func(f *reflect.Value, row *Row) error {
		return fn(f, row)
	}
}

func (fm *cfield) assignUnmarshaller(code int) {
	if code == impsPtr {
		assign(fm, fm.unmarshalPointer)
	} else {
		assign(fm, fm.unmarshalValue)
	}
}

func (fm *cfield) unmarshalPointer(f *reflect.Value, row *Row) error {
	val := row.At(fm.colIndex)
	m := f.Addr().Interface().(Unmarshaler)
	m.UnmarshalCSV(val, row)

	return nil
}

func (fm *cfield) unmarshalValue(f *reflect.Value, row *Row) error {
	val := row.At(fm.colIndex)
	m := f.Interface().(Unmarshaler)
	m.UnmarshalCSV(val, row)
	return nil
}

func (fm *cfield) assignDecoder() {
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

// Sets each field value for the el struct for the given row
func (dec *decoder) set(row *Row, el *reflect.Value) error {
	for _, fm := range dec.fms {
		field := fm.structField

		f := el.FieldByName(field.Name)
		if fm.decode != nil {
			err := fm.decode(&f, row)

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no decoder for %v\n", field.Name)
		}
	}

	return nil
}

func (fm *cfield) decodeBool(f *reflect.Value, row *Row) error {
	val := row.At(fm.colIndex)
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

func (fm *cfield) decodeInt(f *reflect.Value, row *Row) error {
	val := row.At(fm.colIndex)
	i, e := strconv.Atoi(val)

	if e != nil {
		return e
	}

	f.SetInt(int64(i))
	return nil
}

func (fm *cfield) decodeString(f *reflect.Value, row *Row) error {
	val := row.At(fm.colIndex)
	f.SetString(val)

	return nil
}

func (fm *cfield) decodeFloat(bit int) decoderFn {
	return func(f *reflect.Value, row *Row) error {
		val := row.At(fm.colIndex)
		n, err := strconv.ParseFloat(val, bit)

		if err != nil {
			return err
		}

		f.SetFloat(n)

		return nil
	}
}

// ignoreValue does nothing. This is for unsupported types.
func (fm *cfield) ignoreValue(f *reflect.Value, row *Row) error {
	return nil
}
