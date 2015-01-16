package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
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

	return "", fmt.Errorf("no column found for %s", n)
}

type decoder struct {
	csv          *csv.Reader   // the csv document for input
	reflect.Type               // the underlying struct to decode
	out          reflect.Value // the slice output
	fms          []cfield      //
	cols         []string      // colum names
}

// Unmarshaler is the interface implemented by objects which can unmarshall the CSV row itself.
type Unmarshaler interface {

	// UnmarshalCSV receives a string with the column value matching this field
	// and a reference to the the current row.
	// This allows composing a value from mutliple columns.
	UnmarshalCSV(string, *Row) error
}

// Unmarshal parses the CSV document and stores the result in the value pointed to by v. Only a slice of a struct is allowed for v.
//
// The first line of the CSV is document is used for column names.  These are
// paired to matching exported fields in v's type. See Marshal on how to use tags
// to map to different names and additional options.
//
// Supported Types
//
// string, int, float and bool are supported. Any type which implements Unmarshal is also supported.
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
		raw, err := dec.csv.Read()

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
				colIndex:    index,
				structField: f,
			}

			if code, err := impsUnmarshaller(f.Type, new(Unmarshaler)); err == nil {
				fm.assignUnmarshaller(code)
			} else {
				assignDecoder(&fm)
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

		// Get the StructField from the Type
		sf := t.Field(i)

		if skipField(sf) {
			continue
		}

		// Check if the field is CanSet from the value (v)
		if v.Field(i).CanSet() == true {
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

	el := rv.Type().Elem()

	dec := decoder{
		Type: el,
		csv:  r,
		out:  rv,
		cols: cols,
	}

	dec.mapFieldsToCols(el, cols)

	return &dec, nil
}

// Sets each field value for the el struct for the given row
func (dec *decoder) set(row *Row, el *reflect.Value) error {
	for _, fm := range dec.fms {
		field := fm.structField

		f := el.FieldByName(field.Name)
		err := fm.decode(&f, row)

		if err != nil {
			return err
		}
	}

	return nil
}

func assignDecoder(fm *cfield) {
	switch fm.structField.Type.Kind() {
	case reflect.String:
		fm.assign(fm.decodeString)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		fm.assign(fm.decodeInt)
	case reflect.Float32:
		fm.assign(fm.decodeFloat(32))
	case reflect.Float64:
		fm.assign(fm.decodeFloat(64))
	case reflect.Bool:
		fm.assign(fm.decodeBool)
	default:
		fm.assign(fm.ignoreValue)
	}
}
