package websocket

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

// CloseWith is a convenience method to send a CloseMessage, wait for it to be sent, and then close the connection.
// When closeCode is 0, uses CloseNormalClosure.
func (conn *Connection) CloseWith(closeCode int, text string) {
	if closeCode <= 0 {
		closeCode = CloseNormalClosure
	}
	// write the close message
	<-conn.Write(Message{
		Type:  CloseMessage,
		Bytes: websocket.FormatCloseMessage(closeCode, text),
	})

	// and close the connection
	conn.Close()
}

// Close closes the connection to the peer without sending a specific close message.
// See [CloseWith] for providing the client with a reason for the closure.
func (conn *Connection) Close() {
	conn.cancel(errCloseUser)
}

// custom close codes
var (
	errCloseHandlerReturn = &CloseError{Code: websocket.CloseNormalClosure, Text: "Handler close()ed"}
	errCloseHandlerPanic  = &CloseError{Code: websocket.CloseAbnormalClosure, Text: "Handler panic()ed"}
	errCloseOther         = errors.New("unknown close cause")
	errCloseUser          = errors.New("user called close")
)

// CloseError is the cancel cause of [Connection.Context] if a closing handshake too place.
type CloseError = websocket.CloseError

// ReadError is the cancel cause of [Connection.Context] if an error occurred during writing.
type ReadError struct {
	err error
}

func (err ReadError) Error() string {
	return fmt.Sprintf("read error: %v", err.err)
}
func (err ReadError) Unwrap() error {
	return err.err
}

// WriteError is the cancel cause of [Connection.Context] if an error occurred during writing.
type WriteError struct {
	err error
}

func (err WriteError) Error() string {
	return fmt.Sprintf("write error: %v", err.err)
}
func (err WriteError) Unwrap() error {
	return err.err
}
