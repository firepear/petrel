# petrel

Analagous to SQLite's embedding of serverless relational database
capablities within programs, Petrel lets you easily add RPC
capabilities into your programs with no external message broker.

- Optimized for programmer time
- A program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
  - But individual connections are synchronous
- Supports multiple request styles
  - Command-line, with automatic tokenization
  - Blob/JSON, with raw payload pass-thru
- Works over Unix domain sockets or TCP
- Security-conscious design
  - TLS support for link security and/or client authentication
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion,
    accidental or purposeful
- No third-party dependencies
- Proven reliable and decently performant in real-world use

N.B. Petrel is pre-v1.0.0 so assume that there will be breaking changes

## News

* 2022-04-23: v0.33.0: Client and server are separate packages
* 2021-01-09: v0.32.0: More module updates
* 2020-05-14: v0.31.0: Transition to Go module
* 2017-09-30: v0.30.1: Code run through `gofmt -s`
* 2016-11-29: v0.30.0: Internal changes to accomodate petreljs

See the [Release notes](https://github.com/firepear/petrel/raw/master/RELEASE_NOTES) for all updates.

## Documentation

* [Client](https://pkg.go.dev/github.com/firepear/petrel/client?tab=doc) [![GoReportCard link (client)](https://goreportcard.com/badge/github.com/firepear/petrel/client)](https://goreportcard.com/report/github.com/firepear/petrel/client)
* [Server](https://pkg.go.dev/github.com/firepear/petrel/server?tab=doc) [![GoReportCard link (server)](https://goreportcard.com/badge/github.com/firepear/petrel/server)](https://goreportcard.com/report/github.com/firepear/petrel/server)

## Example

See the demo [server](https://github.com/firepear/petrel/blob/master/examples/basic/example-server.go) and
[client](https://github.com/firepear/petrel/blob/master/examples/basic/example-client.go) for
worked examples of Petrel usage.

To see them in action, in the basic example directory, do `go run
example-server.go` to start the example server. Then in another
terminal, try a few runs of the client:

```
go run demo/example-client.go time
go run demo/example-client.go echo whatever you feel like typing here
go run demo/example-client.go
go run demo/example-client.go foobar
```

Check out the results of the client, and the messages printed in the
server's terminal. When you're done, kill the server with `C-c`.
