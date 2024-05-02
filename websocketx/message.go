//spellchecker:words websocketx
package websocketx

//spellchecker:words github gorilla websocket
import (
	"github.com/gorilla/websocket"
)

// Message represents a message sent between client and server.
type Message struct {
	Type int
	Body []byte
}

// Binary checks if this message is a binary message
func (msg Message) Binary() bool {
	return msg.Type == websocket.BinaryMessage
}

// Text checks if this message is a text message
func (msg Message) Text() bool {
	return msg.Type == websocket.TextMessage
}

// NewTextMessage creates a new text message with the given text
func NewTextMessage(text string) Message {
	return Message{
		Type: websocket.TextMessage,
		Body: []byte(text),
	}
}

// NewBinaryMessage creates a new binary message with the given text
func NewBinaryMessage(data []byte) Message {
	return Message{
		Type: websocket.BinaryMessage,
		Body: data,
	}
}

// Prepare prepares a message for sending
func (msg Message) Prepare() (PreparedMessage, error) {
	m, err := websocket.NewPreparedMessage(msg.Type, msg.Body)
	if err != nil {
		return PreparedMessage{}, err
	}
	return PreparedMessage{m: m}, nil
}

// MustPrepare is like Prepare, but panic()s when preparing fails.
func (msg Message) MustPrepare() PreparedMessage {
	m, err := msg.Prepare()
	if err != nil {
		panic(err)
	}
	return m
}

// PreparedMessage represents a message that caches its' on-the-wire encoding.
// This saves re-applying compression.
type PreparedMessage struct {
	m *websocket.PreparedMessage
}
