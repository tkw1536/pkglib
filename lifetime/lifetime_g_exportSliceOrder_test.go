//spellchecker:words lifetime
package lifetime_test

//spellchecker:words pkglib lifetime
import (
	"fmt"

	"go.tkw01536.de/pkglib/lifetime"
)

// RankComponent is anything with a rank method.
type RankComponent interface {
	Component
	Rank() string

	// a special function Rank${TypeName} can be used to sort slices of this type.
	//
	// The method must be named exactly like this, and have a signature func()T where
	// T is of kind float, int or string.
	// Slices are sorted increasingly using the appropriate "<" operator.
	RankRankComponent() int64
}

// Declare two color components Red and Green.

type Captain struct{}

func (Captain) isComponent() {}
func (Captain) RankRankComponent() int64 {
	return 0
}
func (Captain) Rank() string {
	return "Captain"
}

type Admiral struct{}

func (Admiral) isComponent() {}
func (Admiral) RankRankComponent() int64 {
	return 1
}
func (Admiral) Rank() string {
	return "Admiral"
}

// Demonstrates the use of order when exporting slices.
func ExampleLifetime_gExportSliceOrder() {
	// Register components as normal.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Admiral](context)
			lifetime.Place[*Captain](context)
		},
	}

	// export the ranks using ExportSlice.
	// The order is now guaranteed by the RankComponentWeight() function.
	ranks := lifetime.ExportSlice[RankComponent](lt, struct{}{})
	for _, r := range ranks {
		fmt.Println(r.Rank())
	}

	// Output: Captain
	// Admiral
}
