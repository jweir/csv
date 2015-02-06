package csv

import (
	"fmt"
	"reflect"
	"testing"
)

type simple struct {
	Name    string `csv:"FullName"`
	Gender  string
	private int `csv:"-"`
	Age     int
	Ignore  string `csv:"-"`
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

	// Get the header when defined via a tag
	f, _ = x.FieldByName("Ignore")
	_, ok := fieldHeaderName(f)

	if ok == true {
		t.Error("Omitted field returned ok")
	}
}

func TestHeaders(t *testing.T) {
	x := reflect.TypeOf(simple{})

	hh := colNames(x)

	if "[FullName Gender Age]" != fmt.Sprintf("%v", hh) {
		t.Errorf("Incorrected headers: %v", hh)
	}
}
