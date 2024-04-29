# Petrel

_This module is pre-v1; expect frequent breaking changes._

SQLite embeds serverless relational databases into programs. Petrel
lets you do the same with networking and RPC.

- Optimized for programmer time
- A program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
  - But individual connections are synchronous
- Works over Unix domain sockets or TCP
- Security-conscious design
  - TLS support for link security and/or client authentication
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion,
- No external dependencies
- Proven mostly reliable and decently performant in real-world use!

See the [Release
notes](https://github.com/firepear/petrel/blob/main/RELEASE_NOTES.md)
for updates.

[![GoReportCard link (client)](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

# Server

- [Server](https://pkg.go.dev/github.com/firepear/petrel/server?tab=doc)

## Signal handling

Embedding a Petrel server in your code gets you handlers for `SIGINT`
and `SIGTERM`, for free. At the moment, Petrel does not handle
pidfiles.

# Client

- [Client](https://pkg.go.dev/github.com/firepear/petrel/client?tab=doc)
- [Examples](https://github.com/firepear/petrel/raw/main/examples/README.md)

# Protocol

The Petrel wire protocol has a fixed 11-byte header, two run-length
encoded data segments, and an optional 44-byte HMAC segment.

    Status code       uint16 (2 bytes)
    Seqence number    uint32 (4 bytes)
    Request length    uint8  (1 byte)
    Payload length    uint32 (4 bytes)
    ---------------------------------------------------
    Request text      Per request length (max 255 char)
    Payload text      Per payload length (max 4MB)
    ---------------------------------------------------
    HMAC              44 bytes, optional

Request and payload text are utf8 encoded.

There is no need for messages to specify whether HMAC is included or
not, as that is set by the client and server at connection time. HMAC
is base64 utf8 text.
