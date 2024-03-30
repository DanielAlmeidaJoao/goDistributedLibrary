## each protocol has at least two go chans:
    * Message chan: this will only have the messages received
    * error chan: this will only have the errors sent to the protocol, or generic errors

## each protocol will have a map of functions that will handle the messages from those chans

## a server struct will have the following fields:
    * incomming message handler: this will process all the messages received
    * errorHandlerFunction: this will process all the received messages

## TODOs
* Remove a protocol after they have been registered
* Add a protocol after all the others have started
* Should ** CustomConnection ** have an instance of the TCPCHannel ??
* Find a way to gobabel block the main thread
* MSGCODE not being used on the receiver ; only send the client serverAddress if peerTopeer