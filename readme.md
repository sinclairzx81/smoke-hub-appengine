# smoke-hub-appengine

A general purpose messaging relay written in Go for deployment on google's 
appengine infrastructure.

# overview

smoke-hub-appengine is a small messaging / signalling relay for webrtc enabled clients. 
This relay will automatically assign a user a unique address on connection, and allow
that user to forward messages to other addresses on the relay. 

This relay doesn't not advertise accessible addresses of users, and instead relies on 
clients sharing their address outside of the relay. From this relays perspective, 
users are treated anonymously, identified by their address only. 

Although users are anonymous, the relay does support protection from impersonation. Meaning
users are unable to forge their sending address for messages sent through the relay. The end 
result is that if a user receives a message on the relay from another user, they can be sure 
the address is sent from 1 user and 1 user only. 

smoke-hub-appengine is built on top of googles app engine infrastructure. The project
is designed to operate on the google standard environment, which means instances of 
the hub can be deployed for free, and scaled high at cost with 0 configuration.

# documentation

There is none, but running the project locally includes a small test page, and client script
users can use as a example of connecting and leveraging the service.

A test installation can be located at https://smoke-io.appspot.com/.

