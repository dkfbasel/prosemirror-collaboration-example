package websocket

import (
	"encoding/json"
	"time"

	"dkfbasel.ch/orca/collaboration/src/internal/environment"
)

// expire all rooms, that did not receive any action during the given time
const roomExpiration = time.Hour * 5

// WebsocketRoom is used to handle a single editor document
type WebsocketRoom struct {
	Clients    map[*WebsocketClient]bool
	Register   chan *Registration // register a new client
	Unregister chan *Registration // deregister a client

	Notify    chan NotifyMessage // send message to all clients except the sender
	Broadcast chan []byte        // broadcast messages to all clients

	Handler chan Message // handle incoming messages

	DocumentID      string          // unique id of the respective editor content
	DocumentSchema  json.RawMessage // schema of the respective document
	DocumentVersion int64           // current document version on the server
}

// newWebsocketRoom will initialize a new websocket room with corresponding
// handlers
func newWebsocketRoom(srv *environment.Services, id string) *WebsocketRoom {

	room := WebsocketRoom{}
	room.Clients = make(map[*WebsocketClient]bool)
	room.Register = make(chan *Registration)
	room.Unregister = make(chan *Registration)

	// we need a buffer on the notify, to be able to send a message on the
	// notify channel of our room while still handling an incoming message
	room.Notify = make(chan NotifyMessage, 50)

	// we need a buffer on the broadcast, to be able to send a message on the
	// broadcast channel of our room while still handling an incoming message
	room.Broadcast = make(chan []byte, 50)

	// handler for incoming messages
	room.Handler = make(chan Message)

	room.DocumentID = id
	room.DocumentVersion = -1

	// handle registration
	go handleRoom(srv, &room)

	return &room
}
