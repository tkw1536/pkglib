//spellchecker:words sqlx
package sqlx

//spellchecker:words github feiin sqlstring
import (
	"github.com/feiin/sqlstring"
)

// Format formats the provided query with the given parameters.
//
// This function is unsafe on user-controlled input and it should be avoided.
func Format(query string, params ...any) string {
	// NOTE: This function is a wrapper around an external package.
	// As such it is not tested.
	return sqlstring.Format(query, params...)
}
