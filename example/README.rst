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

   ./server            # foreground, log to screen
      or
   ./server &> log.txt # background; log to file

And use the client to send some requests to the server.

When you're done, hit :code:`^C` (if foregrounded) or :code:`kill` (if
backgrounded) to terminate the server cleanly.
