package csv

import (
	"fmt"
	"reflect"
	"testing"
)

func XExampleUnMarshal() {
	type Person struct {
		Name string `csv:"FullName"`
		Age  int
	}

	people := []Person{}

	sample := []byte(`Full Name,Age
John Doe,45`)

	err := Unmarshal(sample, people)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Printf("%+v", people)

	// Output:
	// [{Name: John Doe, Age: 45}]
}

func TestMarshal(t *testing.T) {
	type P struct {
		Name  string
		Age   int
		Happy bool `csv:"Happy" true:"Yes" false:"No"`
	}

	doc := []byte(`Name,Age,ignore,Happy
John,23,,Yes
Jane,27,,No
Bill,28,,Yes`)

	pp := []P{}

	Unmarshal(doc, &pp)

	if len(pp) != 3 {
		t.Errorf("Incorrect decoded length: %d", len(pp))
	}

	for i, v := range []string{"John", "Jane", "Bill"} {
		n := pp[i].Name
		if n != v {
			t.Errorf("expected (%s) got (%s)", v, n)
		}
	}

	for i, v := range []int{23, 27, 28} {
		n := pp[i].Age
		if n != v {
			t.Errorf("expected (%d) got (%d)", v, n)
		}
	}

	for i, v := range []bool{true, false, true} {
		n := pp[i].Happy
		if n != v {
			t.Errorf("expected (%s) got (%s)", v, n)
		}
	}
}

func TestMarshalErrors(t *testing.T) {
	type P struct{}

	doc := []byte(`Name,Age`)

	err := Unmarshal(doc, []P{})

	if err == nil {
		t.Error("No error generated for non-pointer")
	}

	err = Unmarshal(doc, &P{})

	if err == nil {
		t.Error("No error generated for non-slice")
	}

	pp := []P{}

	err = Unmarshal(doc, &pp)

	if err != nil {
		t.Error("Error returned when not expected:", err)
	}
}

func TestPublicFields(t *testing.T) {
	type S struct {
		Name string
		Age  int `csv:"Age"`
		priv int `csv:"-"`
	}

	fs := publicFields(reflect.TypeOf(S{}))

	if len(fs) != 2 {
		t.Error("Incorrect number of exported fields 2 expected got %d", len(fs))
	}

	if fs[0].Name != "Name" || fs[1].Name != "Age" {
		t.Error("Incorrect returned fields")
	}
}

type MFT struct {
	Name    string
	age     string // unexported, should not be included
	Addr    string `csv:"Address"`
	NoMatch int    // public, but no match in the CSV headers
}

func TestMapFields(t *testing.T) {
	pf := publicFields(reflect.TypeOf(MFT{}))

	// 'age' will have no mapping since the age field is not exported
	headers := []string{"Name", "Address", "age"}

	m := mapFields(headers, pf)

	if len(m) != 2 {
		t.Errorf("Incorrect length: %d, %v", len(m), m)
	}

	for i, fn := range []string{"Name", "Address"} {
		if m[fn] != pf[i] {
			t.Errorf("expected %s got %s", fn, pf[i].Name)
		}
	}

}
