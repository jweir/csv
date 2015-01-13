package csv

import (
	"fmt"
)

func ExampleUnmarshal() {
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
