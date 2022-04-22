
# petrel

Petrel is a drop-in package for adding (non-HTTP) network capabilities
to programs.

Analagous to how SQLite embeds the capabilities of relational
databases within programs, Petrel lets you embed message-passing/RPC
capabilities.

Some features:
- Optimized for programmer time
- Proven performant in real-world datacenter use
- Supports command-line style requests with automatic tokenization, like
  ARGV)
- ...Or blob/JSON style request handling (raw payload pass-thru to
  your code)
- Works over Unix domain sockets or TCP
- Security-conscious design:
  - TLS support for link security and/or client authentication
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion,
    accidental or purposeful
- No third-party dependencies
- Passes `golint`, `go vet`, and `go test -race` cleanly
- A program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
  - But each connection is synchronous

## News

[![GoReportCard link](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

* 2022-04-  : v0.33.0: Client and server are separate packages
* 2021-01-09: v0.32.0: More module updates
* 2020-05-14: v0.31.0: Transition to Go module
* 2017-09-30: v0.30.1: Code run through `gofmt -s`
* 2016-11-29: v0.30.0: Internal changes to accomodate petreljs

See the [Release notes](https://github.com/firepear/petrel/raw/master/RELEASE_NOTES) for all updates.

## Documentation

* Install: `go get github.com/firepear/petrel`

* [Package documentation](https://pkg.go.dev/github.com/firepear/petrel/?tab=doc)

## Example

See the demo [server](https://github.com/firepear/petrel/blob/master/examples/basic/example-server.go) and
[client](https://github.com/firepear/petrel/blob/master/examples/basic/example-client.go) for
worked examples of Petrel usage.

To see them in action, in one terminal in the basic example directory,
do `go run example-server.go` to start the example server.

Then in another terminal, try a few runs of the client:

```
go run demo/example-client.go time
go run demo/example-client.go echo whatever you feel like typing here
go run demo/example-client.go
go run demo/example-client.go foobar
```

When you're done, kill the server with `C-c` in its terminal.
