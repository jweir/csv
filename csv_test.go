package csv

import (
	"fmt"
	"reflect"
	"testing"
)

type Person struct {
	Name   string `csv:"FullName"`
	Gender string
	Age    int
	Wallet float32 `csv:"Bank Account"`
}

func ExamplePeople() {
	people := []Person{Person{"Smith, Joe", "M", 23, 19.07}}

	out, _ := Marshal(people)
	fmt.Printf("%s", out)
	// Output:
	// FullName,Gender,Age,Bank Account
	// "Smith, Joe",M,23,19.07
}

type simple struct {
	Name   string `csv:"FullName"`
	Gender string
	Age    int
}

func TestHeader(t *testing.T) {
	x := reflect.TypeOf(simple{})

	// Get the header when defined via a tag
	h, _ := header(x, "Name")

	if h != "FullName" {
		t.Error("header does not match")
	}

	// Use the field FullName when there is no tag
	h, _ = header(x, "Gender")

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
	p := simple{"Jane", "F", 34}

	ty := reflect.ValueOf(p)

	r := encode(ty)

	if fmt.Sprintf("%v", r) != "[Jane F 34]" {
		t.Error("incorrect encoding: %v", r)
	}
}
