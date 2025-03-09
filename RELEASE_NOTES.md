# Release notes

## 0.40.0 (2025-03-xx)

- Signal handling removed from Petrel
- Logging switched to `log/slog`, and custom Loggers can now be passed
  in via `server.Config`
- GenMsg gone in favor of directly pushing Msgs


## 0.39.0 (2025-03-05)

- Servers now have a connection list, providing a handle on Conns
- Conns now have non-numeric ids
- `server.Sockname` -> `server.Addr` since UDP is gone
- Log format update for easier reading and parsing
- Many more tests, which are less brittle than the old tests
- Cleanups/fixes from testing
- Beginning of doc rewrite


## 0.38.0 (2025-02-21)

- Restoration of test suite has begun
- Users no longer have to write two event loops to use Petrel
- Or monitor the chan used for OS signals
- Type `Handler` adds a third return: status (`uint16`)
- Servers now have non-numeric ids (coming to conns soon)
- `Loglvl` moved from `petrel` to `server` and made private
- pre-commit hook now runs lots of checks, and the tests
- `gofmt` and `golangcli-lint` fixups
- Added `toolchain` to `go.mod`


## 0.37.1 (2025-02-21)

- After working on v0.38 for a bit, decided to backport the `gofmt`
  fixups


## 0.37.0 (2025-02-20)

- The wire protocol has changed
  - Protocol version removed
  - Request/transmission status included
- Server and client constructors have been unified
  - UDP connections have been removed
- Server and Client now have `Conn`s which hold state, most especially
  `Resp`onses to requests
  - `Conn` is now passed to net funcs, instead of `server` object
  - `Msgr` code moved from `server` package to `petrel`
  - These two changes together let netcode be better instrumented
- The `Perr` system is gone, replaced with simpler `Status`
- There is now a protocol check/handshake request (`PROTOCHECK`) as
  the first action of every client connection
  - New `Status`: 497, protocol mismatch
- `server.Responder` renamed to `server.Handler`
- `server.Reqlen` renamed to `server.Xferlim`
- `Xferlim` added to `client.Config`
- The `client` related `*Raw` functions have been removed for
  simplicity and operational consistency
- Vastly fewer allocations due to data restructuring


## 0.36.0 (2022-12-18)

- OS signal handling is now managed within Petrel itself. Applications
  can now just watch the channel `server.Sig` rather than having to
  create a channel and link it to `os.Signal` events

## 0.35.0 (2022-12-17)

- The `mode` argument to `server.Register` has been removed
  - As a result, the `server.responder` type has been removed, and the
    server dispatch table now simply maps to functions
- The `argv` mode has been removed. Only what was formerly known as
  `blob` mode is now supported. This has resulted in further changes:
  - `Responder` functions now take `[]byte` instead of `[][]byte`
  - `client.Dispatch` now takes two args: `req, payload []byte` rather
    than the request being the first "chunk" of the payload
- Petrel wire protocol has changed
- `qsplit` dependency has been removed


## 0.34.0 (2022-05-07)

- Msg level `All` is now `Debug`
- Server configuration struct is now named `Config` rather than
  `ServerConfig`
- Client configuration struct is now named `Config` rather than
  `ClientConfig`
- Server configuration now specifies `Msglvl` as a string, which is
  the lowercase version of the Petrel constant (e.g. "debug" for
  `Debug`). This makes application configuration, where the programmer
  shouldn't have to care about Petrel, more sane
- `golint` cleanups


## 0.33.0 (2022-04-23)

- The petrel client and server are now discrete subpackages within the
  module. No code changes other than those needed to make this new
  organization work
- See examples for updated usage; improved docs coming soon


## 0.32.0 (2021-01-09)

- Various other changes to be more in-line with modern Golang project
  expectations


## 0.31.0 (2020-05-14)

- Transition to Go module
- Removal of `petrel.Version`; obviated by module transition
- Renamed `petrel.Protover` to `petrel.Proto`


## 0.30.1 (2017-09-30)

- Code passed through 'goformat -s'. Also a few typo catches and
  similar cleanups.


## 0.30.0 (2016-11-29)

- Internal changes to accomodate petreljs
    - Binary encoding is now little-endian
    - Transmission marshalling has been decoupled from connWrite.
    - connWrite has been split into connWrite and connWriteRaw (the
      former calls the latter to do actual writing of data).
    - connRead has been joined by connReadRaw, which reads from the
      connection without unmarshalling the transmission being
      read. connRead does not call connReadRaw.
    - Client has a new method, DispatchRaw, which forwards
      transmissions to/from a server as-is.


## 0.29.0 (2016-11-20)

- HMAC is now Base64 encoded rather than raw.


## 0.28.0 (2016-11-01)

- The original demo client/server are now at demo/01-basic.
- The basic demo client/server now have a `-hmac` argument to
  allow transmitting with MACs.
- Transmissions now have a sequence number.
- petrel.Client.Dispatch now returns:
    - The transmission response
    - The transmission sequence num
    - An error
- The Petrel wire protocol now includes a 1-byte version identifier. A
  transmission with a version mismatch results in a terminated
  connection, as an HMAC mismatch does.
- Internal names cleanup


## 0.27.0 (2016-10-24)

- ServerConfig.Reqlen is now a uint32 (was int), allowing a maximum
  request size of 4GB.


## 0.26.0 (2016-10-14)

- Petrel now provides optional message authentication via crypto/HMAC
  - Server constructors renamed to match Client ones.
  - Client.Close is now Client.Quit, to match the Server shutdown
    method.
  - OmitPrefix has been removed from ClientConfig. petrel.Client now
    always sends length-prefixed messages.
  - Long manual-style godoc preamble removed in favor of more
    informative documentation for individual structs and methods.
  - Reinstated demo client and server.


## 0.25.0 (2016-10-07)
- Petrel and Pclient packages integrated.
  - Name changes
    - petrel.Config   -> ServerConfig
    - pclient.Config  -> ClientConfig
    - petrel.Handler  -> petrel.Server
    -Handler.AddFunc -> Server.Register
  - All constructors have been renamed (see go doc).
- New option in ServerConfig: Reqlen, the maximum number of bytes
  which will be read for any request.
- New option in ServerConfig: LogIP, whether to log Client IPs on
  connect or not.
- Improvements in reliability and correctness of network handling.
- Client now returns nil and an error when it receives an error
  response from the Server side.
- Better consistency in messaging; efficiency improvements in Msg
  generation.
- Long-form docs are temporarily gutted for rebuild next release.
- Version jump: unified versions and added 1


## 0.22.0 (2016-03-30)
- Use qsplit.LocationsOnce for faster separation of request
  command and args


## 0.21.0 (2016-03-29)
- Name change: asock -> petrel
- asock.Asock is now petrel.Handler
- Asock.AddHandler is now Handler.AddFunc
- Valid settings for Handler.AddFunc's second argument (mode) have
  changed from "split"/"nosplit" to "args"/"blob". This change means
  that the setting now describes its use-case instead of its
  implementation.
- Null request and unknown command messages no longer return a list of
  known handlers. This exposed internal state. Users should implement
  their own 'help' type handler.
- Docs rewrite.


## 0.20.0 (2016-03-09)
- Oneshot connection mode is gone. It was a bad idea, torturously
  implemented.
- Constructors now take -Config instead of Config.
- Improvements to timeout handling (func call eliminated; no more
  local var creation and time.Duration casting on every network
  operation).


## 0.19.1 (2015-12-10)
- Small networking improvements


## 0.19.0 (2015-12-10)
- EOM is gone. All messages are now sent prefixed with their
  length in bytes as a 4 byte header.
- pclient now used where possible for petrel test cases
- Moved example client to the pclient package


## 0.18.0 (2015-08-03)

- Example server now includes a handler which returns an error.
- Test coverage brought back up from rapid 0.16/0.17 dev cycles.
- Documentation improvements


## 0.17.0 (2015-06-18)

- Breaking changes
  - Dispatch is no longer a public type. As a result, it no longer
    appears in constructor calls and does not have to be instantiated
    by the user.
  - DispatchFunc is no longer a public type.  Dispatch functions are
    now added to Asock instances by calling the AddHandler() method.
  - Docs are all kinds of wrong right now. 0.18 will fix this.


## 0.16.1 (2015-06-03)

- petrel/client moved to own package, pclient


## 0.16.0 (2015-06-03)

- Breaking changes
  - Config.TLSConfig no longer exists. It is once again an argument of
    petrel.NewTLS().
  - petrel.NewUnix() has a new argument, which is the Unix
    permissions to set on the socket.
  - Aclient constructors now take a single argument, a Config
    struct, as the Asock constructors do.
  - client.Config has a new `EOM` field, which behaves identically
    to the petrel.Config field of the same name, described below.
  - The Msg generated by a.Quit being invoked (code 199) is now
    sent regardless of Msglvl setting.
  - There is a new Msg, code 401, for a null request.
- Additions and other changes
  - The petrel.Config struct has a new field, `EOM`, which sets the
    end-of-message marker (defaults to "\n\n").
  - Network handling is now more correct and robust.
  - Test suite speedups (across 0.16 and 0.15)


## 0.15.0 (2015-05-26)

- Breaking changes
  - The `Config` struct has two new fields:
    - Config.Buffer specifies the size of the buffer on the
      Asock.Msgr channel (defaults to 32).
    - Config.TLSConfig is the *tls.Config instance which is needed to
      set up a TLS connection.  This was formerly an argument of
      petrel.NewTLS().
    - The `Timeout` field in the `Config` struct is now treated as
      milliseconds instead of seconds.
    - client constructors have a new argument, `timeout`, which sets
      the number of milliseconds before read/write operations timeout.
- Additions and other changes
  - client has a new method, Read()
  - A special-casing of Config.Timeout for test purposes has been
    removed from the production petrel code.
  - Read buffer in petrel and client now defaults to 128 bytes (up
    from 64).


## 0.14.1 (2015-05-19)

- Documentation improvements


## 0.14.0 (2015-05-19)

- TLS support added via the petrel & client NewTLS() methods.


## 0.13.0 (2015-05-16)

- `petrel/client` package added
- Documentation improvements


## 0.12.0 (2015-03-28)

- Msg now implements the error interface, so it will autostringify
  when passed to fmt.Println, log.Print, log.Fatal, etc.


## 0.11.1 (2015-02-18)

- Msg.Conn and Msg.Req are now uint (were int)
- Documentation improvements


## 0.11.0 (2015-02-16)

- Breaking changes
  - Argmode functionality is now per Dispatch function instead of
    per petrel instance.
  - As a result, the signature of Dispatch has changed to
    'map[string]*DispatchFunc', where DispatchFunc is a new type which
    holds information about Dispatch functions.


## 0.10.0 (2015-02-16)


- Breaking changes
  - New Config field, Argmode. This controls how arguments to
    Dispatch functions will be handled.


## 0.9.0 (2015-02-11)

- Breaking changes
  - Dispatch functions are now of type 'func([][]byte) ([]byte error)'
    Formerly the argument was '[]string'.


## 0.8.0 (2015-02-10)

- Breaking changes
  - Constructors now take a configuration struct rather than a long
    list of parameters
  - Unix socket names are no longer semiautomagic; sockets are now
    created exactly where Config.sockname says.


## 0.7.0 (2015-01-23)

- Breaking changes
  - Package name change!
  - New() is now NewUnix()
- Additions and other changes
  - Added TCP support - NewTCP()


## 0.6.0 (2015-01-17)

- Breaking changes
  - New() now has a 4th argument, messaging level, which controls
    which messages will be sent to a.Msgr
- Additions and other changes
  - Msg type is now far more robust, enabling better
    logging. Concommitantly, Msg.Txt is now generally more terse,
    since more information is elsewhere
  - All Msgs now have a numeric Code field, which accurately
    identifies what sort of message it is
  - A Msg is now generated when a response is sent to a client
  - Improved test coverage

- Bugfixes
  - A response is no longer sent when the request generates an error
  - Connection closures due to timeout now generate a Msg indicating
    server-side closure instead of client-side closure


## 0.5.1 (2015-01-10)

- Changes to match new qsplit API


## 0.5.0 (2015-01-05)

- Breaking changes
  - Socket names are no longer automagic. They are now an argument
    to New()
  - Negative timeout values now set a connection deadline in addition
    to creating a one-shot connection (e.g. -5 creates a one-shot conn
    with a 5 second deadline)
  - All dispatches now generate a Msg, not just those which fail
- Additions and other changes
  - sockAccept() and connHandler() are now part of Asock's
    methodset, eliminating several arguments to each
  - There is now a (preëmptive) canonical import path
  - Documentation improvements


## 0.4.1 (2014-12-30)

- Now 'go vet' and golint approved


## 0.4.0 (2014-12-29)

- Public release


## 0.3.0 (2014-10-21)

- Imported code from everydayd

