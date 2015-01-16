package csv

import (
	"fmt"
	"reflect"
	"strconv"
)

type coderFn func(*reflect.Value, *Row) error

// maps a CSV column Name and index to a StructField
type cfield struct {
	colIndex    int
	structField *reflect.StructField

	// this is a function to decode or encode the data, depending on the context
	handler coderFn
}

func (fm *cfield) decode(f *reflect.Value, row *Row) error {
	if fm.handler == nil {
		return fmt.Errorf("no decoder for %v\n", fm.structField.Name)
	}

	return fm.handler(f, row)
}

func (fm *cfield) assign(fn coderFn) {
	fm.handler = func(f *reflect.Value, row *Row) error {
		return fn(f, row)
	}
}

func (fm *cfield) assignUnmarshaller(code int) {
	if code == impsPtr {
		fm.assign(fm.unmarshalPointer)
	} else {
		fm.assign(fm.unmarshalValue)
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

func (fm *cfield) decodeFloat(bit int) coderFn {
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
