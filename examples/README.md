# petrel examples

See the basic
[server](https://github.com/firepear/petrel/blob/main/examples/basic/example-server.go)
and
[client](https://github.com/firepear/petrel/blob/main/examples/basic/example-client.go)
for annotated examples of Petrel usage.

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
