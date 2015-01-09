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
		Name string
		Age  int
	}

	doc := []byte(`Name,Age
John,23
Jane,27
Bill,28`)

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
		age  int `csv:"Age"`
		priv int `csv:"-"`
	}

	fs := publicFields(reflect.TypeOf(S{}))

	if len(fs) != 2 {
		t.Error("Incorrect number of exported fields")
	}

	if fs[0].Name != "Name" || fs[1].Name != "age" {
		t.Error("Incorrect returned fields")
	}
}

type MFT struct {
	Name    string
	age     string `csv:"Age"`
	NoMatch int    // public, but no match in the CSV headers
}

func TestMapFields(t *testing.T) {
	pf := publicFields(reflect.TypeOf(MFT{}))
	headers := []string{"Name", "Age"}
	m := mapFields(headers, pf)

	if len(m) != 2 {
		t.Error("Incorrect length")
	}

	if m["Name"] != pf[0] {
		t.Error("Incorrect mapping")
	}

	if m["Age"] != pf[1] {
		t.Error("Incorrect mapping")
	}

}
