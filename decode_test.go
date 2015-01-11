package csv

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleUnMarshal() {
	type Person struct {
		Name    string `csv:"Full Name"`
		income  string // unexported fields are not Unmarshalled
		Age     int
		Address string `csv:"-"` // skip this field
	}

	people := []Person{}

	sample := []byte(
		`Full Name,income,Age,Address
John Doe,"32,000",45,"125 Maple St"
`)

	err := Unmarshal(sample, &people)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Printf("%+v", people)

	// Output:
	// [{Name:John Doe income: Age:45 Address:}]
}

func TestUnMarshal(t *testing.T) {
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

	fs := exportedFields(reflect.TypeOf(S{}))

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
	rt := reflect.TypeOf(MFT{})

	cols := []string{
		"Name",
		"age", // should not match since the 'age' field is not exported
		"Address",
	}

	fm := mapFieldsToCols(rt, cols)

	if len(fm) != 2 {
		t.Errorf("Expected length of 2, got %d", len(fm))
	}

	for i, n := range []string{"Name", "Address"} {
		if fm[i].colName != n {
			t.Errorf("expected colName of %s got %s", fm[i].colName, n)
		}
	}

	for i, n := range []int{0, 2} {
		if fm[i].colIndex != n {
			t.Errorf("expected colIndex of %d got %d", fm[i].colIndex, n)
		}
	}
}
