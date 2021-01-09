Petrel Multi-HMAC Demo
======================

This demonstration client and server show one way to do per-client
HMAC with Petrel.

The basic mechanism is that the server has one permanent "listener"
petrel.Server instance which only knows how to authenticate connecting
clients and spawn a new Server instance for them to re-connect to. The
full flow goes like this:

    Server                                             Client
    ------------------------------------------------   ---------------------------------------------------
    Listener server (S1), w/TLS starts up
                                                       First petrel.Client (C1, TLS) connects to S1
                                                       C1 sends 'authenticate' request
    S1 checks C1's auth request
    S1 gets C1's HMAC key & generates session ID
    2nd Server (S2, TLS+HMAC) started on new port
    S1 sends response with session ID, port #
                                                       C1 gets successful auth response
                                                       C1 terminates
                                                       2nd Client (C2, TLS+HMAC) connects to S2 on new port
    S2 begins handling requests for C2

After C2 connects to S2, all requests sent will include the session
ID, but this is for internal tracking rather than security. Spoofing
protection is provided by Petrel's HMAC handling: if any request is
received with an HMAC mismatch, the connection is dropped and S2 will
terminate. At that point the client will have to re-connect to S1 and
authenticate again.

This does mean that you can spoof someone if you know their HMAC key
_and_ can MITM their TLS, but at that point all bets are off anyway :)

Inside the server program, new sessions and their petrel.Server
instances are added to a table. Every 3 seconds (so short because this
is an automated demo) this table is scanned and checked for activity;
idle sessions have their petrel.Server instances dropped.

Sending the session ID with each request from C2 lets the Responder
functions of S2 know which entry in the session table they should
update. The first and last thing a Responder should do is update the
last activity time of its session.
