package password

import "fmt"

func ExamplePasswords() {
	// load all the passwords from common sources
	counts := map[string]int{}
	for password := range Passwords(CommonSources()...) {

		// do something with the password
		_ = password.Password // string

		// and in this case count it by source
		counts[password.Source]++
	}

	// output the overall counts!
	fmt.Println(counts)
	// Output: map[common/top10_000.txt:10000]
}
