//spellchecker:words sqlx
package sqlx

//spellchecker:words strings unicode
import (
	"strings"
	"unicode"
)

// IsSafeDatabaseSingleQuote checks if value can safely be put inside 's inside a database query.
func IsSafeDatabaseSingleQuote(value string) bool {
	return !strings.ContainsAny(value, "'`") // TODO: This should be safer, but it's relatively controlled
}

// IsSafeDatabaseLiteral checks if a value is safe to be used as a database query literal.
func IsSafeDatabaseLiteral(value string) bool {
	// the empty name is not allowed!
	if len(value) == 0 {
		return false
	}

	// reserved words aren't allowed!
	if _, reserved := reservedSQLWords[strings.ToLower(value)]; reserved {
		return false
	}

	letters := []rune(value)

	// the first letter must be a unicode letter, a @, _ or #.
	if !unicode.IsLetter(letters[0]) && letters[0] != '@' && letters[0] != '_' && letters[0] != '#' {
		return false
	}

	// each subsequent letter may be a unicode letter, a unicode number, @, _, # or $.
	for _, l := range letters[1:] {
		if !unicode.IsLetter(l) && !unicode.IsNumber(l) && l != '@' && l != '_' && l != '-' && l != '#' && l != '$' {
			return false
		}
	}

	return true
}

// reservedSQLWords is a list of restricted sql words.
var reservedSQLWords = map[string]struct{}{
	//spellchecker:disable-next-line
	"absolute": {}, "action": {}, "ada": {}, "add": {}, "admin": {}, "after": {}, "aggregate": {}, "alias": {}, "all": {}, "allocate": {}, "alter": {}, "and": {}, "any": {}, "are": {}, "array": {}, "as": {}, "asc": {}, "asensitive": {}, "assertion": {}, "asymmetric": {}, "at": {}, "atomic": {}, "authorization": {}, "avg": {}, "backup": {}, "before": {}, "begin": {}, "between": {}, "binary": {}, "bit": {}, "bit_length": {}, "blob": {}, "boolean": {}, "both": {}, "breadth": {}, "break": {}, "browse": {}, "bulk": {}, "by": {}, "call": {}, "called": {}, "cardinality": {}, "cascade": {}, "cascaded": {}, "case": {}, "cast": {}, "catalog": {}, "char": {}, "character": {}, "character_length": {}, "char_length": {}, "check": {}, "checkpoint": {}, "class": {}, "clob": {}, "close": {}, "clustered": {}, "coalesce": {}, "collate": {}, "collation": {}, "collect": {}, "column": {}, "commit": {}, "completion": {}, "compute": {}, "condition": {}, "connect": {}, "connection": {}, "constraint": {}, "constraints": {}, "constructor": {}, "contains": {}, "containstable": {}, "continue": {}, "convert": {}, "corr": {}, "corresponding": {}, "count": {}, "covar_pop": {}, "covar_samp": {}, "create": {}, "cross": {}, "cube": {}, "cume_dist": {}, "current": {}, "current_catalog": {}, "current_date": {}, "current_default_transform_group": {}, "current_path": {}, "current_role": {}, "current_schema": {}, "current_time": {}, "current_timestamp": {}, "current_transform_group_for_type": {}, "current_user": {}, "cursor": {}, "cycle": {}, "data": {}, "database": {}, "date": {}, "day": {}, "dbcc": {}, "deallocate": {}, "dec": {}, "decimal": {}, "declare": {}, "default": {}, "deferrable": {}, "deferred": {}, "delete": {}, "deny": {}, "depth": {}, "deref": {}, "desc": {}, "describe": {}, "descriptor": {}, "destroy": {}, "destructor": {}, "deterministic": {}, "diagnostics": {}, "dictionary": {}, "disconnect": {}, "disk": {}, "distinct": {}, "distributed": {}, "domain": {}, "double": {}, "drop": {}, "dump": {}, "dynamic": {}, "each": {}, "element": {}, "else": {}, "end": {}, "end-exec": {}, "equals": {}, "errlvl": {}, "escape": {}, "every": {}, "except": {}, "exception": {}, "exec": {}, "execute": {}, "exists": {}, "exit": {}, "external": {}, "extract": {}, "false": {}, "fetch": {}, "file": {}, "fillfactor": {}, "filter": {}, "first": {}, "float": {}, "for": {}, "foreign": {}, "fortran": {}, "found": {}, "free": {}, "freetext": {}, "freetexttable": {}, "from": {}, "full": {}, "fulltexttable": {}, "function": {}, "fusion": {}, "general": {}, "get": {}, "global": {}, "go": {}, "goto": {}, "grant": {}, "group": {}, "grouping": {}, "having": {}, "hold": {}, "holdlock": {}, "host": {}, "hour": {}, "identity": {}, "identitycol": {}, "identity_insert": {}, "if": {}, "ignore": {}, "immediate": {}, "in": {}, "include": {}, "index": {}, "indicator": {}, "initialize": {}, "initially": {}, "inner": {}, "inout": {}, "input": {}, "insensitive": {}, "insert": {}, "int": {}, "integer": {}, "intersect": {}, "intersection": {}, "interval": {}, "into": {}, "is": {}, "isolation": {}, "iterate": {}, "join": {}, "key": {}, "kill": {}, "language": {}, "large": {}, "last": {}, "lateral": {}, "leading": {}, "left": {}, "less": {}, "level": {}, "like": {}, "like_regex": {}, "limit": {}, "lineno": {}, "ln": {}, "load": {}, "local": {}, "localtime": {}, "localtimestamp": {}, "locator": {}, "lower": {}, "map": {}, "match": {}, "max": {}, "member": {}, "merge": {}, "method": {}, "min": {}, "minute": {}, "mod": {}, "modifies": {}, "modify": {}, "module": {}, "month": {}, "multiset": {}, "names": {}, "national": {}, "natural": {}, "nchar": {}, "nclob": {}, "new": {}, "next": {}, "no": {}, "nocheck": {}, "nonclustered": {}, "none": {}, "normalize": {}, "not": {}, "null": {}, "nullif": {}, "numeric": {}, "object": {}, "occurrences_regex": {}, "octet_length": {}, "of": {}, "off": {}, "offsets": {}, "old": {}, "on": {}, "only": {}, "open": {}, "opendatasource": {}, "openquery": {}, "openrowset": {}, "openxml": {}, "operation": {}, "option": {}, "or": {}, "order": {}, "ordinality": {}, "out": {}, "outer": {}, "output": {}, "over": {}, "overlaps": {}, "overlay": {}, "pad": {}, "parameter": {}, "parameters": {}, "partial": {}, "partition": {}, "pascal": {}, "path": {}, "percent": {}, "percentile_cont": {}, "percentile_disc": {}, "percent_rank": {}, "pivot": {}, "plan": {}, "position": {}, "position_regex": {}, "postfix": {}, "precision": {}, "prefix": {}, "preorder": {}, "prepare": {}, "preserve": {}, "primary": {}, "print": {}, "prior": {}, "privileges": {}, "proc": {}, "procedure": {}, "public": {}, "raiserror": {}, "range": {}, "read": {}, "reads": {}, "readtext": {}, "real": {}, "reconfigure": {}, "recursive": {}, "ref": {}, "references": {}, "referencing": {}, "regr_avgx": {}, "regr_avgy": {}, "regr_count": {}, "regr_intercept": {}, "regr_r2": {}, "regr_slope": {}, "regr_sxx": {}, "regr_sxy": {}, "regr_syy": {}, "relative": {}, "release": {}, "replication": {}, "restore": {}, "restrict": {}, "result": {}, "return": {}, "returns": {}, "revert": {}, "revoke": {}, "right": {}, "role": {}, "rollback": {}, "rollup": {}, "routine": {}, "row": {}, "rowcount": {}, "rowguidcol": {}, "rows": {}, "rule": {}, "save": {}, "savepoint": {}, "schema": {}, "scope": {}, "scroll": {}, "search": {}, "second": {}, "section": {}, "securityaudit": {}, "select": {}, "sensitive": {}, "sequence": {}, "session": {}, "session_user": {}, "set": {}, "sets": {}, "setuser": {}, "shutdown": {}, "similar": {}, "size": {}, "smallint": {}, "some": {}, "space": {}, "specific": {}, "specifictype": {}, "sql": {}, "sqlca": {}, "sqlcode": {}, "sqlerror": {}, "sqlexception": {}, "sqlstate": {}, "sqlwarning": {}, "start": {}, "state": {}, "statement": {}, "static": {}, "statistics": {}, "stddev_pop": {}, "stddev_samp": {}, "structure": {}, "submultiset": {}, "substring": {}, "substring_regex": {}, "sum": {}, "symmetric": {}, "system": {}, "system_user": {}, "table": {}, "tablesample": {}, "temporary": {}, "terminate": {}, "textsize": {}, "than": {}, "then": {}, "time": {}, "timestamp": {}, "timezone_hour": {}, "timezone_minute": {}, "to": {}, "top": {}, "trailing": {}, "tran": {}, "transaction": {}, "translate": {}, "translate_regex": {}, "translation": {}, "treat": {}, "trigger": {}, "trim": {}, "true": {}, "truncate": {}, "tsequal": {}, "uescape": {}, "under": {}, "union": {}, "unique": {}, "unknown": {}, "unnest": {}, "unpivot": {}, "update": {}, "updatetext": {}, "upper": {}, "usage": {}, "use": {}, "user": {}, "using": {}, "value": {}, "values": {}, "varchar": {}, "variable": {}, "varying": {}, "var_pop": {}, "var_samp": {}, "view": {}, "waitfor": {}, "when": {}, "whenever": {}, "where": {}, "while": {}, "width_bucket": {}, "window": {}, "with": {}, "within": {}, "without": {}, "work": {}, "write": {}, "writetext": {}, "xmlagg": {}, "xmlattributes": {}, "xmlbinary": {}, "xmlcast": {}, "xmlcomment": {}, "xmlconcat": {}, "xmldocument": {}, "xmlelement": {}, "xmlexists": {}, "xmlforest": {}, "xmliterate": {}, "xmlnamespaces": {}, "xmlparse": {}, "xmlpi": {}, "xmlquery": {}, "xmlserialize": {}, "xmltable": {}, "xmltext": {}, "xmlvalidate": {}, "year": {}, "zone": {}}
