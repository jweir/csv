package csv

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleMarshal() {
	type Person struct {
		Name    string `csv:"FullName"`
		Gender  string
		Age     int
		Wallet  float32 `csv:"Bank Account"`
		Happy   bool    `true:"Yes!" false:"Sad"`
		private int     `csv:"-"`
	}

	people := []Person{
		Person{
			Name:   "Smith, Joe",
			Gender: "M",
			Age:    23,
			Wallet: 19.07,
			Happy:  false,
		},
	}

	out, _ := Marshal(people)
	fmt.Printf("%s", out)
	// Output:
	// FullName,Gender,Age,Bank Account,Happy
	// "Smith, Joe",M,23,19.07,Sad
}

func TestMarshal_without_a_slice(t *testing.T) {
	_, err := Marshal(simple{})

	if err == nil {
		t.Error("Non slice produced no error")
	}

}

type simple struct {
	Name    string `csv:"FullName"`
	Gender  string
	private int `csv:"-"`
	Age     int
}

func TestHeader(t *testing.T) {
	x := reflect.TypeOf(simple{})

	// Get the header when defined via a tag
	f, _ := x.FieldByName("Name")
	h, _ := header(f)

	if h != "FullName" {
		t.Error("header does not match")
	}

	// Use the field FullName when there is no tag
	f, _ = x.FieldByName("Gender")
	h, _ = header(f)

	if h != "Gender" {
		t.Error("Default header FullName not created")
	}
}

func TestHeaders(t *testing.T) {
	x := reflect.TypeOf(simple{})

	hh := headers(x)

	if "[FullName Gender Age]" != fmt.Sprintf("%v", hh) {
		t.Errorf("Incorrected headers: %v", hh)
	}
}

func TestEncode(t *testing.T) {
	p := simple{
		Name:    "Jane",
		Gender:  "F",
		Age:     34,
		private: 1,
	}

	ty := reflect.ValueOf(p)

	r := encode(ty)

	if fmt.Sprintf("%v", r) != "[Jane F 34]" {
		t.Error("incorrect encoding: %v", r)
	}
}

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
