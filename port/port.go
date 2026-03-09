//spellchecker:words port
package port

//spellchecker:words context errors time
import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	errFreePortRange = errors.New("free port is not positive")
	errFindFreePort  = errors.New("failed to find free port")
	errCloseListener = errors.New("failed to close listener")
)

// FindFreePort picks a random (positive) unassigned port on the given host.
// It is only guaranteed to be free at the time the function is invoked, and other programs may race to claim it after the function returns.
// If no free port is found, or ctx expires, an error is returned.
func FindFreePort(ctx context.Context, host string) (int, error) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", host+":0")
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errFindFreePort, err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, fmt.Errorf("%w: %w", errCloseListener, err)
	}
	if port <= 0 {
		return 0, errFreePortRange
	}
	return port, nil
}

const DefaultWaitPortInterval = 10 * time.Millisecond

// WaitForPort repeatedly attempts to connect to the given tcp address until a connection succeeds.
// Once the connection succeeds, it is immediately closed.
//
// interval determines the duration between connection attempts, if zero [DefaultWaitPortInterval] is used.
//
// If the context closes before a connection is successful, returns an error wrapping the context error.
func WaitForPort(ctx context.Context, addr string, interval time.Duration) error {
	var dialer net.Dialer

	if interval <= 0 {
		interval = DefaultWaitPortInterval
	}

	timer := time.NewTimer(interval)
	for {
		var conn net.Conn
		var err error

		// try to connect with the given timeout
		conn, err = dialer.DialContext(ctx, "tcp", addr)

		// if we connected close the connection again
		if err == nil {
			if err := conn.Close(); err != nil {
				return fmt.Errorf("failed to close connection: %w", err)
			}
			return nil
		}

		// wait to try again, or close if the context is done.
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-timer.C:
		}
	}
}
