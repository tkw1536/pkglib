//spellchecker:words lifetime
package lifetime_test

//spellchecker:words reflect pkglib lifetime
import (
	"fmt"
	"reflect"

	"go.tkw01536.de/pkglib/lifetime"
)

// Shape is a component type with a RankShape() method to rank it.
type Shape interface {
	RankShape() string
}

type Square struct{}

func (Square) RankShape() string {
	return "0"
}

type Circle struct{}

func (Circle) RankShape() string {
	return "1"
}

type Triangle struct{}

func (Triangle) RankShape() string {
	return "2"
}

// Demonstrates that sorting slices applies to the Component type also.
// See also Example G.
func ExampleLifetime_hAllOrder() {
	// Create a lifetime with the square, circle and triangle components
	lt := &lifetime.Lifetime[Shape, struct{}]{
		Register: func(context *lifetime.Registry[Shape, struct{}]) {
			lifetime.Place[*Square](context)
			lifetime.Place[*Circle](context)
			lifetime.Place[*Triangle](context)
		},
	}

	// Retrieve all components and print their names.
	// The order is guaranteed to be consistent here.
	for _, shape := range lt.All(struct{}{}) {
		name := reflect.TypeOf(shape).Elem().Name()
		fmt.Println(name)
	}

	// Output: Square
	// Circle
	// Triangle
}
