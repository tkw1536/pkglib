//spellchecker:words status
package status

//spellchecker:words github pkglib stream
import (
	"fmt"
	"io"

	"github.com/tkw1536/pkglib/stream"
)

// WriterGroup intelligently runs handler over items concurrently.
//
// Count determines the number of concurrent invocations to run.
// count <= 0 indicates no limit.
// count = 1 indicates running handler in order.
//
// handler is additionally passed a writer.
// When there is only one concurrent invocation, the original writer as a parameter.
// When there is more than one concurrent invocation, each invocation is passed a single line of a new [Status].
// The [Status] will send output to the standard output of str.
//
// WriterGroup returns the first non-nil error returned by each call to handler; or nil otherwise.
func WriterGroup[T any](writer io.Writer, count int, handler func(value T, output io.Writer) error, items []T, opts ...StreamGroupOption[T]) error {

	// create a group
	var group Group[T, error]
	group.HandlerLimit = count

	// apply all the options
	isParallel := count != 1
	for _, opt := range opts {
		group = opt(isParallel, group)
	}

	// setup the default prefix string
	if group.PrefixString == nil {
		group.PrefixString = DefaultPrefixString[T]
	}

	// then just iterate over the items
	if !isParallel {
		for index, item := range items {
			fmt.Fprintln(writer, group.PrefixString(item, index))
			err := handler(item, writer)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// if we are running in parallel, setup a handler
	group.Handler = func(item T, index int, writer io.Writer) error {
		return handler(item, writer)
	}

	// create a new status display
	st := NewWithCompat(writer, 0)
	st.Start()
	defer st.Stop()

	// and use it!
	return UseErrorGroup(st, group, items)
}

// StreamGroupOption represents an option for [WriterGroup].
// The boolean indicates if the option is being applied to a status line or not.
//
// NOTE: This name is here for backwards compatibility reasons.
type StreamGroupOption[T any] func(bool, Group[T, error]) Group[T, error]

// SmartMessage sets the message to display as a prefix before invoking a handler.
func SmartMessage[T any](handler func(value T) string) StreamGroupOption[T] {
	return func(p bool, s Group[T, error]) Group[T, error] {
		s.PrefixString = func(item T, index int) string {
			message := handler(item)
			if p {
				return "[" + message + "]: "
			}
			return message
		}
		s.PrefixAlign = true
		return s
	}
}

// StreamGroup is like WriterGroup, but operates on an IOStream.
//
// When underlying operations are non-interactive, use WriterGroup instead.
func StreamGroup[T any](str stream.IOStream, count int, handler func(value T, str stream.IOStream) error, items []T, opts ...StreamGroupOption[T]) error {
	return WriterGroup(str.Stdout, count, func(value T, output io.Writer) error {
		var gstr stream.IOStream
		if output != str.Stdout {
			gstr = stream.NonInteractive(output)
		} else {
			gstr = str
		}
		return handler(value, gstr)
	}, items, opts...)
}
