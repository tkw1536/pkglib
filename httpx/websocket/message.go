package websocket

// Message represents a message sent between client and server.
type Message struct {
	Type  int
	Bytes []byte
}

// Binary checks if this message is a binary message
func (msg Message) Binary() bool {
	return msg.Type == BinaryMessage
}

// Text checks if this message is a text message
func (msg Message) Text() bool {
	return msg.Type == TextMessage
}
