# Petrel

The current release (v0.38) builds on 0.37's reworking of the codebase
and is bringing the test suite back up to par. However, even more
testing work (and benchmarks!) are coming in 0.39, as well as some
further internal rework to make using Petrel simpler for programmers.

Please wait until then to use Petrel.

----

SQLite lets programmers add reliable, structured data storage to any
program, with no hassle, low resource overhead, and no infrastructure
requirements. Petrel aims to do the same for networking.

- Optimized for programmer time
- A single program can embed multiple Petrel servers and/or clients
- Petrel servers support arbitrarily many concurrent connections
- Security-conscious design
  - TLS support for link security
  - HMAC support for message verification
  - Message length limits to protect against memory exhaustion
- No non-stdlib dependencies
- Proven mostly reliable and decently performant in real-world use!

See the [Release
notes](https://github.com/firepear/petrel/blob/main/RELEASE_NOTES.md)
for info on updates.

![Build/test status](https://github.com/firepear/petrel/actions/workflows/go.yml/badge.svg)
[![GoReportCard link](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

# Using Petrel

For people whose learning style might be described as "code first,
then start reading when something breaks", here are links to the
[server](https://pkg.go.dev/github.com/firepear/petrel/server?tab=doc)
and
[client](https://pkg.go.dev/github.com/firepear/petrel/client?tab=doc)
godoc. You might find reading over the
[example](https://github.com/firepear/petrel/raw/main/examples/README.md)
code helpful as well -- the `server` and `client` there are very well
commented.

For people who want more documentation as an introduction before
loading a buffer into the editor, please keep reading.

- [Server](#server)
  - [OS signals](#os-signals)
  - [Handlers](#handlers)
- [Client](#client)
- [Statuses](#statuses)
- [Protocol](#protocol)
- [Code quality](#code-quality)

## Basic concepts

The goal of Petrel is not to replace message brokers, pub/sub systems,
or libraries like gRPC -- just as SQLite does not try to replace
high-performance RDBMSes.

Speaking as its author, What I want from Petrel is for it to be there
every time I have a thought like "Man, I'd like to be able to talk to
this over the network" or "It would be great to have a C&C channel
into this long-running process". I strive to be make it decently
useful, performant, and correct, but its origins are fundamentally as
a _sysadmin's tool._

- Petrel has a client/server model, in which clients `Dispatch()`
  requests to servers, and get responses in reply
- Petrel pushes raw bytes over sockets; there is no underlying library
  or service handling network traffic
- Petrel offers network security, but it does not have any concept of
  _authentication;_ that is an application-level concern
- Petrel is very unopinionated from the perspective of plugging into
  it:
  - It does not care what your data looks like; interaly everything is
    a `[]byte`, and it's up to your application to know what to do
    with the payload of a given request
  - Request handlers are just functions with the signature
    `func([]byte) (uint16, []byte, error)`
- On the other hand, it's kinda opinionated about operations (on the
  server side), in the name of taking stuff off your plate:
  - You'll need to have a simple event loop somewhere
  - You get a free handler for `SIGINT` and `SIGTERM` (whether you
    wanted it or not)


## Server


### OS signals

Embedding a Petrel server in your code gets you handlers for `SIGINT`
and `SIGTERM`, for free. Petrel does not handle pidfiles or other
aspects of daemonization.

### Writing Handlers

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

# Code quality

This is a one-person project, but I take it seriously and do my best
to deliver code that is well-tested and does what I mean it to do. At
the shallow end, these badges let you know that (1) the `main` branch
builds cleanly, and (2) the project structure is sane:

![Build/test status](https://github.com/firepear/petrel/actions/workflows/go.yml/badge.svg)
[![GoReportCard link](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

And then going a bit deeper, there's the [test coverage
report](https://firepear.github.io/petrel/assets/coverage.html), which
is generated by my `pre-commit` hook. And speaking of, that hook also
runs

- `gofmt`
- `go vet`
- `golangci-lint`
- `staticcheck`

So I can't make a single commit if my tests or those tools see a
failure. That doesn't mean no bugs, but it does mean the code is free
of a lot of bad smells, as well as any bugs that I've seen before.

## Running tests

If you want to run the tests yourself, either just run
`./assets/runcover` or look at it to see how I'm running the tests --
they will fail if run with just `go test ./...`

