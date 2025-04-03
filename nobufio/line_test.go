//spellchecker:words nobufio
package nobufio_test

//spellchecker:words strings github pkglib nobufio
import (
	"fmt"
	"strings"

	"github.com/tkw1536/pkglib/nobufio"
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
