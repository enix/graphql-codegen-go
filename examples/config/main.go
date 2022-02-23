package main

import (
	"fmt"
)

func main() {
	person := internal.Person{
		FirstName: "",
		Lastname:  "",
		Age:       nil,
		Gender:    nil,
		Address:   nil,
	}
	fmt.Println(person)
}
