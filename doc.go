// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

/*
Package petrel implements an embeddable and programmer-friendly
mechanism for RPC. The goal is to be to networking what SQLite is to
backing datastores.

The package itself contains documentation as well as code, data, and
types that are shared between petrel/server and petrel/client. Those
packages should be imported rather than this one.

# Usage

# Wire protocol

The Petrel wire protocol has a fixed 10-byte header, an optional 32
byte HMAC segment, and two run-length encoded segments.

    Seqence number    uint32 (4 bytes)
    Protocol version  uint8  (1 byte)
    Request length    uint8  (1 byte)
    Payload length    uint32 (4 bytes)
    ------------------------------------
    Request text      Per request length
    Payload text      Per payload length
    ------------------------------------
    HMAC              32 bytes, optional

There is no need for wire messages to specify whether HMAC is included
or not, as that is negotiated between the client and server when the
connection is made.

*/
package petrel
