//spellchecker:words websocketx
package websocketx

//spellchecker:words compress flate time github gorilla websocket
import (
	"compress/flate"
	"time"

	"github.com/gorilla/websocket"
)

// Options describes options for connections made to a [Server].
// It is not guaranteed that options changed after the first call to [ServeHTTP] are accepted.
//
// See [Options.SetDefault] for appropriate defaults.
type Options struct {
	// HandshakeTimeout specifies the duration for the open and close handshakes to complete.
	// If the handshakes fail within this time, the connection is closed immediately with an error.
	HandshakeTimeout time.Duration

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then buffers allocated by the HTTP server are used. The
	// I/O buffer sizes do not limit the size of the messages that can be sent
	// or received.
	ReadBufferSize, WriteBufferSize int

	// ReadInterval is the maximum amount of time to wait for a new packet on
	// the underlying network connection.
	// After a read has timed out, the websocket connection state is corrupt and
	// the connection will be closed.
	//
	// To ensure that the client sends at least some packet within this time, use [PingInterval].
	ReadInterval time.Duration

	// WriteInterval is the write deadline on the underlying network
	// connection. After a write has timed out, the websocket state is corrupt and
	// the connection will be closed.
	WriteInterval time.Duration

	// WriteBufferPool is a pool of buffers for write operations. If the value
	// is not set, then write buffers are allocated to the connection for the
	// lifetime of the connection.
	//
	// A pool is most useful when the application has a modest volume of writes
	// across a large number of connections.
	//
	// Applications should use a single pool for each unique value of
	// WriteBufferSize.
	WriteBufferPool websocket.BufferPool

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this field is not nil, then the server negotiates a
	// subprotocol by selecting the first match in this list with a protocol
	// requested by the client. If there's no match, then no protocol is
	// negotiated (the Sec-Websocket-Protocol header is not included in the
	// handshake response).
	Subprotocols []string

	// PingInterval is the amount of time between successive Ping messages sent to be sent to the peer.
	// A ping message is intended to invoke a Poke response from the client, ensuring that the connection
	// has not been corrupted, and a package is received in at most [ReadInterval].
	//
	// For this purpose, PingInterval should be less than [ReadInterval].
	PingInterval time.Duration

	// ReadLimit is the maximum size in bytes for a message read from the peer. If a
	// message exceeds the limit, the connection sends a close message to the peer
	// and returns an error to the application.
	ReadLimit int64

	// CompressionLevel is the compression level of packages received to and from the peer.
	// See the [compress/flate] package for compression levels.
	//
	// A compression level of [flate.NoCompression] means that compression is disabled.
	// This is ignored if compression has not been negotiated with the client.
	CompressionLevel int
}

// CompressionEnabled determines if the server should attempt to negotiate per
// message compression (RFC 7692). This method returning true does not
// guarantee that compression will be supported; only that the server will
// attempt to negotiate enabling it.
//
// This is enabled when the CompressionLevel is not set to [flate.NoCompression].
func (opt *Options) CompressionEnabled() bool {
	return opt.CompressionLevel != flate.NoCompression
}

// Defaults for [Options].
const (
	DefaultWriteInterval    = 10 * time.Second
	DefaultReadInterval     = time.Minute
	DefaultHandshakeTimeout = time.Second
	DefaultPingInterval     = (DefaultReadInterval * 9) / 10
	DefaultReadLimit        = 2048 // bytes
)

// SetDefaults sets defaults for options.
// See the appropriate default constants for this package.
func (opt *Options) SetDefaults() {
	if opt.WriteInterval == 0 {
		opt.WriteInterval = DefaultWriteInterval
	}
	if opt.ReadInterval == 0 {
		opt.ReadInterval = DefaultReadInterval
	}
	if opt.PingInterval <= 0 {
		opt.PingInterval = DefaultPingInterval
	}
	if opt.ReadLimit <= 0 {
		opt.ReadLimit = DefaultReadLimit
	}
	if opt.HandshakeTimeout <= 0 {
		opt.HandshakeTimeout = DefaultHandshakeTimeout
	}
}
