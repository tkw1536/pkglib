package wrap

import "fmt"

func ExampleString() {
	fmt.Println(String(20, "this text is longer than 20 characters and is wrapped accordingly"))
	fmt.Println(String(20, "	(wrapping also takes leading spaces into account)"))
	// Output: this text is longer
	// than 20 characters
	// and is wrapped
	// accordingly
	//	(wrapping also
	//	takes leading
	//	spaces into
	//	account)
}
