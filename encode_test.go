package csv

import (
	"reflect"
	"testing"
	"bytes"
)

type X struct {
	First string
}

type P struct {
	First string
	Last  string
}

func (p P) MarshalCSV() ([]byte, error) {
	return []byte(p.First + " " + p.Last), nil
}

func TestMarshal_without_a_slice(t *testing.T) {
	_, err := Marshal(simple{})

	if err == nil {
		t.Error("Non slice produced no error")
	}
}

func TestEncodeFieldValue(t *testing.T) {
	var encTests = []struct {
		val      interface{}
		expected string
		tag      string
	}{
		// Strings
		{"ABC", "ABC", ""},
		{byte(123), "123", ""},

		// Numerics
		{int(1), "1", ""},
		{float32(3.2), "3.2", ""},
		{uint32(123), "123", ""},
		{complex64(1 + 2i), "(+1+2i)", ""},

		// Boolean
		{true, "Yes", `true:"Yes" false:"No"`},
		{false, "No", `true:"Yes" false:"No"`},

		// TODO Array
		// Interface with Marshaler
		{P{"Jay", "Zee"}, "Jay Zee", ""},

		// Struct without Marshaler will produce nothing
		{X{"Jay"}, "", ""},
	}

	enc := &encoder{}

	for _, test := range encTests {
		fv := reflect.ValueOf(test.val)
		st := reflect.StructTag(test.tag)
		res := enc.encodeCol(fv, st)

		if res != test.expected {
			t.Errorf("%s does not match %s", res, test.expected)
		}
	}

}

func TestMarshalCsvOfStructs(t *testing.T) {

	type ST struct {
		A string
	}

	expected := []byte(`A
ASD
a
` )
	data := []ST{{"ASD"}, {"a"}}
	out, err := Marshal(data)
	if err != nil {
		t.Logf("Failed to marshal to csv ")
		t.Fail()
	}
	if ! bytes.Equal(out, expected) {
		t.Logf("Failed to marshal to correct format")
		t.Fail()
	}
}

func TestMarshalCsvOfPointers(t *testing.T) {

	expected := []byte(`A
ASD
` )
	tPointer := []interface{}{}
	tPointer = append(tPointer, struct{ A string }{"ASD"})
	out, err := Marshal(tPointer)

	if err != nil {
		t.Logf("Failed to marshal to csv ")
	}

	if ! bytes.Equal(out, expected) {
		t.Logf("Failed to marshal to correct format")
		t.Fail()
	}
}
