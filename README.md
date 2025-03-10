# Petrel

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

For people who want more exposition before loading a buffer into the
editor, please keep reading.

- [The basics](#the-basics)
- [Servers](#servers)
  - [Handlers](#handlers)
  - [Status](#status)
  - [Monitoring](#monitoring)
- [Clients](#clients)
- [Protocol](#protocol)
- [Code quality](#code-quality)

## The basics

The goal of Petrel is not to replace message brokers, pub/sub systems,
or libraries like gRPC -- just as SQLite does not try to replace
high-performance RDBMSes.

Speaking as its author, What I want from Petrel is for it to be there
every time I have a thought like "Man, I'd like to be able to talk to
this over the network" or "It would be great to have a C&C channel
into this long-running process". I strive to be make it decently
useful, performant, and correct, but its origins are fundamentally as
a _sysadmin's tool._ It's glue code for the network.

- Petrel has a client/server model, in which clients `Dispatch()`
  requests to servers, and get responses in reply
- Petrel pushes raw bytes over sockets; there is no underlying library
  or service handling network traffic
- Petrel offers network security, but it does not have any concept of
  _authentication;_ that is an application-level concern
- Petrel is very unopinionated from the perspective of fitting into
  your code:
  - It does not care what your data looks like; interally everything is
    a `[]byte`, and it's up to your application to know what to do
    with the payload of a given request
  - Request handlers are just functions with the signature
    `func([]byte) (uint16, []byte, error)`
  - You can change where and how a Petrel server logs messages by
    passing in a custom `slog.Logger`


## Servers

A Petrel server wants to be hands-off. Once a server is instantiated
and handlers have been configured, your code should be able to ignore
it during normal operation.

The simplest case of creating a server is:

```
import ps "github.com/firepear/petrel/server"

s, err := ps.New(&ps.Config{Addr: "localhost:60606"})
if err != nil {
        // take care of oops here
}
```

At this point `s` is a live `server`, which is listening for
connections but doesn't know how to do anything else. To make it
useful we need to add at least one `Handler`, which we'll cover in the
next section.

To shut a `server` down, call `s.Quit()` and all background details
will be taken care of as it sweeps up behind itself.

### Handlers

A `Handler` is a functions which a `server` calls to _handle_ a
request. To make a `Handler` callable, we `Register()` them. If we
have a function named `startsWith` which returns words from a list
that begin with a given letter, we could call

`s.Register("matches", startsWith)`

And now `s` knows how to respond to clients that send `matches`
requests. Two things of note:

- `Handler` funcs do not need to be exported so long as they are in
  the package where the `server` is instantiated
- The registered label for a request (`matches` in this example) has
  no relation to the function name

A `Handler` has the signature `func([]byte) (uint16, []byte, error)`

The input and return `[]byte` are the request and response payloads,
respectively. The other two return values are more nuanced.

When writing a `Handler`, the `error` value should always be `nil`
unless the function experiences an actual and unrecoverable error. To
put it another way, the `error` is not a signal to the client that the
request was non-normal; it is a signal to the `server` that the
`Handler` function has failed.

The `uint16` is the response status, and is covered in more detail in
the next section.

### Status

Request and response status are actually part of the Petrel wire
protocol header, and they are the primary mechanism for communicating
how things are going between a client and server.

Status is represented as a `uint16`, so it takes up two bytes and has
a range of possible values from 0 (zero) to 65535. Petrel reserves the
range 0-1024 for system use.

This does not mean that `Handler` code should never return a status of
0-1024 -- in fact, the standard response status for "success" is
`200`, just as in HTTP. What it means is that an application should
never (re)define the meaning of a status in that range. The list of
currently-defined statuses can be found in the `petrel` package godoc.

Let's look at a couple of examples. In the previous section we
mentioned a hypothetical `Handler` func named `startsWith` which takes
a single letter as an input and returns a list of words which start
with that letter. Maybe its complete list of words to check is really
small:

`["apple", "boy", "bat", "dog"]`

If a client sends a request for words starting with `a`, we might
return

`return 200, []byte("apple"), nil`

But if a client requests words starting with `n`, there are no
matches. This is not an error condition from the `Handler` point of
view; the code executed successfully, but there is nothing to
return. We might indicate this by returning status 200 and an empty
payload:

`return 200, []byte{}, nil`

But this feels ambiguous. It might be better to design our app such
that a given status indicates "no matches" and we use the payload to
provide more information:

`return 9101, []byte(fmt.Sprintf("no matches for %s"), letter), nil`

We might want something with even more structure, and design a JSON
struct which can encode all the information our app might need to
convey for any given response. That's easy, since Petrel treats all
payloads as slices of bytes and leaves it up to application code to
parse and/or assemble those bytes. But that's getting off-topic.


### Monitoring

If `Handler` funcs almost always return `nil` for their `error` value,
and application code is not involved in the dispatch of requests and
responses, how can we know what's going on with the server?

First, the server generates log messages, as you might expect.  By
default messages log to `stderr`, at a log level of Debug.  You can
control which messages are logged and where they go by creating a
custom `*slog.Logger` and passing it in as `server.Config.Logger`. See
[log/slog](https://pkg.go.dev/log/slog) for details.

Second, the server exports a `chan error` called `Shutdown`. When a
`server` encounters a shutdown condition a single message will be sent
over this channel. You should watch it in order to know when/if
something happens to the server. If your application doesn't already
have an event loop or other construct for monitoring channels, then
something like this might be a starting point:

```
keepalive := true
for keepalive {
	select {
	case msg := <-s.Shutdown:
		log.Println("app event loop:", msg)
		keepalive = false
		break
	}
}
```

This is taken directly from `examples/server/basic-server.go`, where
you can see it with many comments to explain what's going on. But the
important thing is to keep an eye on `s.Shutdown` so that you can take
appropriate steps when a `Msg` shows up there. The specifics are up to
you, but `s.Quit` will have already been called, so there's no need to
bother with that.

## Clients

Petrel clients are very lightweight. There is no concept of a
long-lived client which open and close multiple network connections
(or a single connection multiple times). A client can handle any
number of requests, but the design intent is that once you call
`Quit()` to drop the connection, you are done with that client and
will create a new one if needed.

Here is a minimal case of a client:

```
import pc "github.com/firepear/petrel/client"

c, err := pc.New(&pc.Config{Addr: sn})
// handle err
err = c.Dispatch("foo", []byte{SOME_PAYLOAD})
// handle err
if c.Resp.Status == 200 {
    fmt.Println(string(c.Resp.Payload))
}
c.Quit()
```

We instantiate a `client`, and send a `foo` request via
`Dispatch()`. If there was no error, we check the status in the
response struct (`c.Resp`) and then print the returned payload if the
status indicated success.

Check out `examples/client/basic-client` for a longer example, with
many more comments.

## Network security

TLS and HMAC functionality are in place, but are currently untested
and undocumeted following the v0.37 rewrite. The next release (v0.40)
will add tests and full documentation for them. For now, please refer
to the godoc.

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

I do my best to deliver code that is well-tested and does what I mean
it to do. At the shallow end of demonstrating that, these badges let
you know that most recent tag builds cleanly, and the project
structure is sane:

![Build/test status](https://github.com/firepear/petrel/actions/workflows/go.yml/badge.svg)
[![GoReportCard link](https://goreportcard.com/badge/github.com/firepear/petrel)](https://goreportcard.com/report/github.com/firepear/petrel)

Going deeper, there's the [test coverage
report](https://firepear.github.io/petrel/assets/coverage.html), which
is generated by my `pre-commit` hook... which also runs

- `gofmt`
- `go vet`
- `golangci-lint`
- `staticcheck`

So no commits happen if any tests fail or those tools flag an
issue. That doesn't mean no bugs, but it does mean the code is free of
a lot of bad smells, as well as any bugs that I've seen before.

## Running tests

If you want to run the tests yourself, either just run
`./assets/runcover` or look at it to see how I'm running the tests --
they will fail if run with a bare `go test ./...` due to relying on
code which is conditionally compiled in only for the tests.
