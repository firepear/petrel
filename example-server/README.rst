Install petrel and pclient.

::

   go get firepear.net/petrel
   go get firepear.net/pclient

Build and launch the server in one terminal.

::

   $ cd $GOPATH/src/firepear.net/petrel/example-server && go build
   $ ./example-server # run in foreground. kill with ^c to quit.

In another terminal, build and launch the client.

::

   $ cd $GOPATH/src/firepear.net/pclient/example-client && go build
   $ ./example-client # will provide list of known commands
