//spellchecker:words validator
package validator

//spellchecker:words errors strconv
import (
	"errors"
	"fmt"
	"strconv"
)

// Demonstrates a passing validation.
func ExampleValidate() {
	var value struct {
		Number    int    `validate:"positive" default:"234"`
		String    string `validate:"nonempty" default:"stuff"`
		Recursive struct {
			Number int    `validate:"positive" default:"45"`
			String string `validate:"nonempty" default:"more"`
		} `recurse:"true"`
	}

	// Create a validator collection
	collection := make(Collection, 2)

	// positive checks if a number is positive
	Add(collection, "positive", func(Value *int, Default string) error {
		// if value is unset, parse the default as a string
		if *Value == 0 {
			i, err := strconv.ParseInt(Default, 10, 64)
			if err != nil {
				return err
			}
			*Value = int(i)
			return nil
		}

		// check that we are actually positive!
		if *Value < 0 {
			return errors.New("not positive")
		}
		return nil
	})

	// nonempty checks that a string is not empty
	Add(collection, "nonempty", func(Value *string, Default string) error {
		// set the default
		if *Value == "" {
			*Value = Default
		}

		// check that it is not empty
		if *Value == "" {
			return errors.New("empty string")
		}
		return nil
	})

	err := Validate(&value, collection)
	fmt.Printf("%v\n", value)
	fmt.Println(err)
	// Output: {234 stuff {45 more}}
	// <nil>
}

// Demonstrates a failing validation
func ExampleValidate_fail() {

	// Create a validator collection
	collection := make(Collection, 2)

	// positive checks if a number is positive
	Add(collection, "positive", func(Value *int, Default string) error {
		// if value is unset, parse the default as a string
		if *Value == 0 {
			i, err := strconv.ParseInt(Default, 10, 64)
			if err != nil {
				return err
			}
			*Value = int(i)
			return nil
		}

		// check that we are actually positive!
		if *Value < 0 {
			return errors.New("not positive")
		}
		return nil
	})

	// nonempty checks that a string is not empty
	Add(collection, "nonempty", func(Value *string, Default string) error {
		// set the default
		if *Value == "" {
			*Value = Default
		}

		// check that it is not empty
		if *Value == "" {
			return errors.New("empty string")
		}
		return nil
	})

	// declare a value that uses the validators
	var value struct {
		Number    int    `validate:"positive" default:"12"`
		String    string `validate:"nonempty" default:"stuff"`
		Recursive struct {
			Number int    `validate:"positive" default:"12"`
			String string `validate:"nonempty"`
		} `recurse:"true"`
	}

	err := Validate(&value, collection)

	fmt.Printf("%v\n", value)
	fmt.Println(err)
	// Output: {12 stuff {12 }}
	// field "Recursive": field "String": empty string
}

// Demonstrates that Validate cannot be called on a non-struct type.
func ExampleValidate_notAStruct() {
	var value int
	err := Validate(&value, nil)

	fmt.Println(err)
	// Output: validate called on non-struct type
}

// Demonstrates that non-validators cause an error.
func ExampleValidate_notAValidator() {

	// create a collection with something that isn't a validator
	collection := make(Collection, 2)
	collection["generic"] = "I_AM_NOT_A_VALIDATOR"

	// try to validate a field with a non-validator
	var value struct {
		Field int `validate:"generic"`
	}
	err := Validate(&value, collection)

	fmt.Println(err)

	// Output: field "Field": entry "generic" in validators is not a validator
}

// Demonstrates that validator types are checked.
func ExampleValidate_invalid() {

	// create a collection with a string validator
	collection := make(Collection, 2)
	collection["string"] = func(Value *string, Default string) error {
		panic("never reached")
	}

	// try to validate an int field with the incompatible validator
	var value struct {
		Field int `validate:"string"`
	}
	err := Validate(&value, collection)

	fmt.Println(err)
	// Output: field "Field": validator "string": got type string, expected type int
}
