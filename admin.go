// Adminsock provides a Unix domain socket
package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

const (
	VERSION = "0.4.0"
)

func New(map[string]func([]string) ([]byte, error)) (chan bool, error) {
	// TODO after proving this out, define an interface of a custom
	// type based on the current signature and then change signature
	// to map[string]CustomType
	q, err := launchListener()        // create listener
	return q, err
}
