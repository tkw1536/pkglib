//spellchecker:words password
package password_test

//spellchecker:words pkglib password
import (
	"fmt"

	"go.tkw01536.de/pkglib/password"
)

func ExamplePasswords() {
	// load all the passwords from common sources
	counts := map[string]int{}
	for pass := range password.Passwords(password.CommonSources()...) {
		// do something with the password
		_ = pass.Password // string

		// and in this case count it by source
		counts[pass.Source]++
	}

	// output the overall counts!
	fmt.Println(counts)
	// Output: map[common/top10_000.txt:10000]
}
