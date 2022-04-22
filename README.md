
# petrel

Petrel is a non-HTTP toolkit for adding network capabilities to
programs. With it you can define APIs/RPCs of arbitrary complexity,
using any data format you like. Here are some key features:

- Optimized for programmer ease-of-use, but has been proven decently
  performant in real-world datacenter use.

- Supports command-line style requests (automatic tokenization, like
  ARGV), or blob style request handling (raw payload passed through to
  your code).

- Works over Unix domain sockets or TCP.

- Security-conscious design:

  - TLS support for link security and/or client authentication.

  - HMAC support for message verification.

  - Message length limits to protect against memory exhaustion,
    accidental or purposeful

- Petrel has no external dependencies, and passes `golint`,
  `go vet`, and `go test -race` cleanly.

## News

[![GoReportCard link](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

* 2020-05-14: v0.31.0: Transition to Go module
* 2017-09-30: v0.30.1: Code run through `gofmt -s`
* 2016-11-29: v0.30.0: Internal changes to accomodate petreljs
* 2016-11-20: v0.29.0: HMAC is now base64 encoded
* 2016-11-01: v0.28.0: HMAC improvements; added protocol version

See the [Release notes](https://github.com/firepear/petrel/raw/master/RELEASE_NOTES) for all updates.

## Documentation

* Install: `go get github.com/firepear/petrel`

* [Package documentation](https://pkg.go.dev/github.com/firepear/petrel/?tab=doc)

See the demo [server](https://github.com/firepear/petrel/blob/master/demo/01-basic/server.go) and
[client](https://github.com/firepear/petrel/blob/master/demo/01-basic/client.go) for
worked examples of Petrel usage.

To see them in action, in one terminal, do `'go run example-server.go'` to start the example
server.

In another terminal, try a few runs of the client, like::

```
go run demo/example-client.go time
go run demo/example-client.go echo whatever you feel like typing here
go run demo/example-client.go
go run demo/example-client.go foobar
```

When you're done, kill the server with `C-c` in its terminal.
