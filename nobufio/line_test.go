package nobufio

import (
	"fmt"
	"strings"
)

func ExampleReadLine() {
	input := strings.NewReader("line1\nline2\r\n\n\r\nline5")

	fmt.Println(ReadLine(input))
	fmt.Println(ReadLine(input))
	fmt.Println(ReadLine(input))
	fmt.Println(ReadLine(input))
	fmt.Println(ReadLine(input))
	fmt.Println(ReadLine(input))
	// Output: line1 <nil>
	// line2 <nil>
	//  <nil>
	//  <nil>
	// line5 <nil>
	//  EOF
}
