//spellchecker:words lreflect
package lreflect_test

//spellchecker:words reflect pkglib lifetime interal lreflect
import (
	"fmt"
	"reflect"

	"go.tkw01536.de/pkglib/lifetime/interal/lreflect"
)

//spellchecker:words rankable

type RankableStruct string

// That can be used to sort the slice by length.
func (r RankableStruct) RankRankableStruct() uint64 {
	return uint64(len(r))
}

// That can be used to sort the slice in inverted length fashion.
func (r RankableStruct) RankRankableInterface() int {
	return -len(r)
}

type RankableInterface interface {
	// RankRankableInterface sorts slices of type RankableInterface
	RankRankableInterface() int
}

func ExampleSortSliceByRank() {
	{
		// take a slice of type RankableStruct, and sort by RankRankableStruct
		values := []RankableStruct{
			"yoda",
			"am",
			"i",
		}

		_ = lreflect.SortSliceByRank(reflect.ValueOf(values))
		fmt.Println(values)
	}

	{
		// take a slice of type RankableInterface, and sort by RankRankableInterface
		values := []RankableInterface{
			RankableStruct("i"),
			RankableStruct("yoda"),
			RankableStruct("am"),
		}
		_ = lreflect.SortSliceByRank(reflect.ValueOf(values))

		fmt.Println(values)
	}

	{
		// take a slice of type string, and don't sort it (because no sort method exists)
		values := []string{
			"i",
			"yoda",
			"am",
		}
		_ = lreflect.SortSliceByRank(reflect.ValueOf(values))

		fmt.Println(values)
	}

	// Output: [i am yoda]
	// [yoda am i]
	// [i yoda am]
}
