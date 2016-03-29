Install petrel and pclient.

::

   go get firepear.net/petrel
   go get firepear.net/pclient

To play with the example client and server, build them.

::
   
   cd $GOPATH/src/firepear.net/petrel/example && go build server.go
   cd $GOPATH/src/firepear.net/pclient/example && go build client.go

Launch the server in one terminal.

::

   $GOPATH/src/firepear.net/petrel/example/server # run in foreground
                                                 # kill with ^c

In another terminal, experiment with the client.

::

   $GOPATH/src/firepear.net/client/example/client # will provide list of
                                                  # known commands
