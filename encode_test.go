package csv

import (
	"reflect"
	"testing"
)

type P struct {
	First string
	Last  string
}

func (p P) MarshalCSV() ([]byte, error) {
	return []byte(p.First + " " + p.Last), nil
}

type X struct {
	First string
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

	for _, test := range encTests {
		fv := reflect.ValueOf(test.val)
		st := reflect.StructTag(test.tag)
		res := encodeFieldValue(fv, st)

		if res != test.expected {
			t.Errorf("%s does not match %s", res, test.expected)
		}
	}

}
