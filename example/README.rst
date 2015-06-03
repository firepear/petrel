Install asock and aclient.

::

   go get firepear.net/asock
   go get firepear.net/aclient

To play with the example client and server, build them.

::
   
   go build server.go
   go build client.go

Then launch the server.

::
   
   ./server &> log.txt

And use the client to send some requests to the server.

When you're done, use :code:`kill` to terminate the server cleanly,
then check out the logfile.
