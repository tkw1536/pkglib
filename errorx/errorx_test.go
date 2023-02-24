// Package errorx provides First
package errorx

import (
	"errors"
	"fmt"
)

func ExampleFirst() {
	fmt.Println(First(errors.New("something"), errors.New("something else")))
	fmt.Println(First(nil, nil, errors.New("something"), nil, errors.New("something else")))
	fmt.Println(First(nil))
	// Output: something
	// something
	// <nil>
}
