package csv

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type P struct {
		String    string
		Int       int
		unxported int
		Bool      bool `true:"Yes" false:"No"`
		Float32   float32
		Float64   float64
		Complex64 complex64 `csv:"C64"`
		// Struct
		// Interface
		// Array
	}

	doc := []byte(`String,Int,unexported,Bool,Float32,Float64,C64
John,23,1,Yes,32.2,64.1,1
Jane,27,2,No,33.1,65.1,2
Bill,28,3,Yes,34.7,65.1,3`)

	pp := []P{}

	Unmarshal(doc, &pp)

	if len(pp) != 3 {
		t.Errorf("Incorrect record length: %d", len(pp))
	}

	assert := func(e, a interface{}) {
		if e != a {
			t.Errorf("expected (%s) got (%s)", e, a)
		}
	}

	strs := []string{"John", "Jane", "Bill"}
	ints := []int{23, 27, 28}
	bools := []bool{true, false, true}
	f32s := []float32{32.2, 33.1, 34.7}
	f64s := []float64{64.1, 65.1, 65.1}

	for i, p := range pp {
		assert(strs[i], p.String)
		assert(ints[i], p.Int)
		assert(bools[i], p.Bool)
		assert(f32s[i], p.Float32)
		assert(f64s[i], p.Float64)
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
