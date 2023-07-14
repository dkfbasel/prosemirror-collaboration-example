package websocket

import (
	"dkfbasel.ch/orca/collaboration/src/internal/environment"
)

// handleRoom will handle all messages sent to the given room
func handleRoom(srv *environment.Services, room *WebsocketRoom) {

	// run a loop to handle all incoming messages for this room
	for {
		select {
		// register a client in a document room
		case registration := <-room.Register:
			room.Clients[registration.Client] = true
			registration.Client.MessageHandler = room.Handler
			close(registration.Done)

		// unregister a client from a document room
		case registration := <-room.Unregister:
			delete(room.Clients, registration.Client)
			close(registration.Done)

		// broadcast a message to all clients (including sender)
		case message := <-room.Broadcast:
			for client := range room.Clients {
				client.Send <- message
			}

		// notify all other clients in the room (excluding sender)
		case notify := <-room.Notify:
			for client := range room.Clients {
				if client != notify.Client {
					client.Send <- notify.Payload
				}
			}

		// handle incoming messages
		case message := <-room.Handler:
			handleMessage(srv, room, &message)
		}
	}
}
