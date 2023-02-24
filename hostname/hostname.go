// Package hostname provides the hostname.
package hostname

import (
	"os"

	"github.com/Showmax/go-fqdn"
)

// FQDN returns the best attempt at a fully qualified domain name of the host system.
//
// This is basically a thing wrapper around "github.com/Showmax/go-fqdn".
// It falls back to ["os".Hostname] in case of an error.
func FQDN() string {
	// NOTE(twiesing): Not entirely how to test this.
	// Considering porting the entire package.

	// first try to use the hostname function
	{
		name, err := fqdn.FqdnHostname()
		if err == nil {
			return name
		}
	}

	// then fall back to the hostname (or "" if that also fails)
	{
		hostname, err := os.Hostname()
		if err != nil {
			return ""
		}
		return hostname
	}
}
