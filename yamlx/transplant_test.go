//spellchecker:words yamlx
package yamlx_test

//spellchecker:words github pkglib yamlx
import (
	"fmt"

	"github.com/tkw1536/pkglib/yamlx"
)

func ExampleTransplant() {
	template := mustUnmarshal(nil, "person:\n    # the name of the person\n    name: \"name goes here\"\n    # how they see the dress\n    dress_color: \"e.g. black and blue\"\n")

	// the actual values
	values := mustUnmarshal(nil, "person:\n    name: \"Testy Tester\"\n    dress_color: \"black and blue\"\n")

	// "transplant", i.e. copy over all the values
	if err := yamlx.Transplant(template, values, false); err != nil {
		panic(err)
	}

	fmt.Println(mustMarshal(nil, template))
	// Output: person:
	//     # the name of the person
	//     name: "Testy Tester"
	//     # how they see the dress
	//     dress_color: "black and blue"
}

func ExampleReplace() {
	// some people see the dress and white and gold
	dress := mustUnmarshal(nil, "dress:\n    color: \"white and gold\"")

	// others as blue and black
	blue_and_black, err := yamlx.Marshal("blue and black")
	if err != nil {
		panic(err)
	}

	// replace the color of the dress with blue and black
	if err := yamlx.Replace(dress, *blue_and_black, "dress", "color"); err != nil {
		panic(err)
	}

	fmt.Println(mustMarshal(nil, dress))
	// Output: dress:
	//     color: blue and black
}
