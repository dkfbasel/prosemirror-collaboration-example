package websocket

import (
	"dkfbasel.ch/orca/collaboration/src/internal/environment"
	"dkfbasel.ch/orca/pkg/logger"
)

// Hub for socket handling
type WebsocketHub struct {
	Srv        *environment.Services     // external services
	Rooms      map[string]*WebsocketRoom // room for documents
	Register   chan *Registration        // register a new client
	Unregister chan *Registration        // deregister a client
}

// newHub will create a new websocket hub to handle various document rooms
// as well as registration and unregistration of clients
func newHub(srv *environment.Services) *WebsocketHub {

	hub := WebsocketHub{}
	hub.Rooms = make(map[string]*WebsocketRoom)
	hub.Register = make(chan *Registration)
	hub.Unregister = make(chan *Registration)
	hub.Srv = srv

	go func() {
		for {
			select {
			// register a client in the hub
			case registration := <-hub.Register:

				room, ok := hub.Rooms[registration.Client.DocumentID]
				if !ok {
					room = newWebsocketRoom(srv, registration.Client.DocumentID)
					hub.Rooms[registration.Client.DocumentID] = room
				}

				room.Register <- registration

			case registration := <-hub.Unregister:
				room, ok := hub.Rooms[registration.Client.DocumentID]
				if !ok {
					continue
				}

				// save the document id to possibly remove the
				// whole room at a later point
				documentID := registration.Client.DocumentID

				room.Unregister <- registration
				<-registration.Done

				// note: no registration can occur until this case is
				// finished, therefore we do not need to lock our rooms

				// remove the room if there are no more clients
				if len(room.Clients) == 0 {
					logger.Debug("remove room", logger.String("room", documentID))
					delete(hub.Rooms, documentID)
				}
			}
		}
	}()

	return &hub

}
