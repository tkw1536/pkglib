package sshx

import (
	"bytes"
	"unicode"

	"golang.org/x/crypto/ssh"
)

// ParseKeys repeatedly parses a PublicKey in authorized_keys format from in.
// At most limit keys are parsed, where limit < 0 indicates an unlimited number of keys to be parsed.
//
// If parsing fails, all previously parsed keys, comments and options are returned along with a non-nil error.
// If there are fewer keys than limit, an error is returned.
func ParseKeys(in []byte, limit int) (keys []ssh.PublicKey, comments []string, options [][]string, rest []byte, err error) {
	// shortcut: zero keys to be parsed
	if limit == 0 {
		return nil, nil, nil, in, nil
	}

	// allocate a set of outputs
	if limit > 0 {
		keys = make([]ssh.PublicKey, 0, limit)
		comments = make([]string, 0, limit)
		options = make([][]string, 0, limit)
	} else {
		limit = -1
	}

	// a set of temporary variables
	// holding the last parsed item
	var key ssh.PublicKey
	var comment string
	var option []string

	for rest = in; limit != 0 && len(rest) != 0; limit-- {
		// parse from the remaining input
		key, comment, option, rest, err = ssh.ParseAuthorizedKey(rest)
		if err != nil {
			break
		}

		// append the parsed data
		keys = append(keys, key)
		comments = append(comments, comment)
		options = append(options, option)

		// trim all leading and trailing spaces
		rest = bytes.TrimLeftFunc(rest, unicode.IsSpace)
	}

	return
}

// ParseAllKeys parses all keys in authorized_keys format.
//
// Parsing stops if the input is exhausted, or when an equivalent ParseKeys call would return an error.
// This function exists for convenience; use ParseKeys for more fine-grained control.
func ParseAllKeys(in []byte) (keys []ssh.PublicKey) {
	// NOTE: We could use ParseKeys() from above
	// but inlining saves a bunch of memory

	var key ssh.PublicKey
	var err error

	for {
		key, _, _, in, err = ssh.ParseAuthorizedKey(in)
		if err != nil {
			break
		}

		// append the parsed data
		keys = append(keys, key)
	}

	return keys
}
