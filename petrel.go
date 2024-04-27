// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/signal"
	"syscall"
)

const (
	// Proto is the version of the wire protocol implemented by
	// this library
	Proto = uint8(0)
)

var (
	pverbuf = new(bytes.Buffer)
	Sigchan chan os.Signal
)

func init() {
	// pre-compute the binary encoding of Proto
	binary.Write(pverbuf, binary.LittleEndian, Proto)

	// we'll listen for SIGINT and SIGTERM so we can behave like a
	// proper service. (mostly; we're not writing out a pidfile.)
	// we need a channel to receive signals on.
	Sigchan = make(chan os.Signal, 1)
	// and we need to register that channel to listen for the
	// signals we want.
	signal.Notify(Sigchan, syscall.SIGINT, syscall.SIGTERM)
	// we will now respond properly to 'kill' calls to our pid,
	// and to C-c at the terminal we're running in.
}
