// Adminsock provides a Unix domain socket
package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

const (
	VERSION = "0.4.0"
)

// Dispatcher
type Dispatcher map[string]func ([]string) ([]byte, error)

// New returns two channels. The first is a write-only channel which
// will shut down the socket and all connections on write. The second
// is a read-only channel which will deliver errors from the socket.
func New(d Dispatcher) (chan bool, chan error, error) {
	return launchListener()        // create listener
}
