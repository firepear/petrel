# Petrel

_This code is pre-v1; expect frequent breaking changes._

SQLite easily embeds a serverless relational database into any
program. Petrel does the same for non-HTTP networking. Highlights:

- Optimized for programmer time
- A single program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
  - But individual connections are synchronous
- Security-conscious design
  - TLS support for link security
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion
- No dependencies outside the Go standard library
- Proven mostly reliable and decently performant in real-world use!

See the [Release
notes](https://github.com/firepear/petrel/blob/main/RELEASE_NOTES.md)
for updates.

![Build/test status](https://github.com/firepear/petrel/actions/workflows/go.yml/badge.svg)
[![GoReportCard link (client)](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

# Using Petrel

This is the reference implementation, which is written in Golang. I
have plans to write client implementations in Python and Swift in the
near future. They'll be linked here when they exist.

For a minimal introduction, see the
[server](https://pkg.go.dev/github.com/firepear/petrel/server?tab=doc)
and/or
[client](https://pkg.go.dev/github.com/firepear/petrel/client?tab=doc)
godoc, and check out the
[examples](https://github.com/firepear/petrel/raw/main/examples/README.md).

For a more leisurely walkthrough, keep reading.

## Server


### Signal handling

Embedding a Petrel server in your code gets you handlers for `SIGINT`
and `SIGTERM`, for free. (Petrel does not handle pidfiles or other
aspects of daemonization.)

## Client


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

HMAC is base64 utf8 text. There is no need for messages to specify
whether HMAC is included or not, as that is set by the client and
server at connection time.
