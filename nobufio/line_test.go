//spellchecker:words nobufio
package nobufio_test

//spellchecker:words strings pkglib nobufio
import (
	"fmt"
	"strings"

	"go.tkw01536.de/pkglib/nobufio"
)

func ExampleReadLine() {
	input := strings.NewReader("line1\nline2\r\n\n\r\nline5")

	fmt.Println(nobufio.ReadLine(input))
	fmt.Println(nobufio.ReadLine(input))
	fmt.Println(nobufio.ReadLine(input))
	fmt.Println(nobufio.ReadLine(input))
	fmt.Println(nobufio.ReadLine(input))
	fmt.Println(nobufio.ReadLine(input))
	// Output: line1 <nil>
	// line2 <nil>
	//  <nil>
	//  <nil>
	// line5 <nil>
	//  EOF
}
