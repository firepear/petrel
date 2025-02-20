# petrel examples

See the basic
[server](https://github.com/firepear/petrel/blob/main/examples/server/basic-server.go)
and
[client](https://github.com/firepear/petrel/blob/main/examples/client/basic-client.go)
for annotated examples of Petrel usage.

To see them in action, in the basic server directory, do `go run
basic-server.go` to start the example server. Then in another
terminal, try a few runs of the client:

```
go run examples/client/basic-client.go time
go run examples/client/basic-client.go echo whatever you feel like typing here
go run examples/client/basic-client.go
go run examples/client/basic-client.go foobar
```

Check out the results of the client, and the messages printed in the
server's terminal. When you're done, kill the server with `C-c`.
