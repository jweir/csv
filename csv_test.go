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
	h, _ := fieldHeaderName(f)

	if h != "FullName" {
		t.Error("header does not match")
	}

	// Use the field FullName when there is no tag
	f, _ = x.FieldByName("Gender")
	h, _ = fieldHeaderName(f)

	if h != "Gender" {
		t.Error("Default header FullName not created")
	}
}

func TestHeaders(t *testing.T) {
	x := reflect.TypeOf(simple{})

	hh := typeHeaders(x)

	if "[FullName Gender Age]" != fmt.Sprintf("%v", hh) {
		t.Errorf("Incorrected headers: %v", hh)
	}
}
