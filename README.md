# Petrel

_This module is pre-v1; breaking changes will be flagged in release notes_

SQLite embeds serverless relational databases into programs. Petrel
does the same for networking and RPC.

- Optimized for programmer time
- A program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
  - But individual connections are synchronous
- Works over Unix domain sockets or TCP
- Security-conscious design
  - TLS support for link security and/or client authentication
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion,
    accidental or purposeful
- No external dependencies (Go stdlib only)
- Proven mostly reliable and decently performant in real-world use!

See the [Release
notes](https://github.com/firepear/petrel/blob/main/RELEASE_NOTES.md)
for updates.

[![GoReportCard link (client)](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

## API Documentation

- [Client](https://pkg.go.dev/github.com/firepear/petrel/client?tab=doc)
- [Server](https://pkg.go.dev/github.com/firepear/petrel/server?tab=doc)
- [Examples](https://github.com/firepear/petrel/raw/main/examples/README.md)

## Over the wire

The Petrel wire protocol has a fixed 10-byte header, two run-length
encoded data segments, and an optional 44-byte HMAC segment.

    Seqence number    uint32 (4 bytes)
    Protocol version  uint8  (1 byte)
    Request length    uint8  (1 byte)
    Payload length    uint32 (4 bytes)
    ------------------------------------
    Request text      Per request length
    Payload text      Per payload length
    ------------------------------------
    HMAC              44 bytes, optional

There is no need for wire messages to specify whether HMAC is included
or not, as that is negotiated between the client and server at
connection time.

