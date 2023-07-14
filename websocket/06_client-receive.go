package websocket

import (
	"context"
	"encoding/json"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"dkfbasel.ch/orca/pkg/logger"
	"nhooyr.io/websocket"
)

func handleReceive(hub *WebsocketHub, client *WebsocketClient) error {
	// handle incoming requests
	for {
		// read data from the socket
		_, dta, err := client.Conn.Read(context.Background())

		// handle closing of websockets
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			logger.Debug("socket closed normally")

			// unregister the client from the hub
			registration := newRegistration(client)
			hub.Unregister <- registration
			return nil
		}

		// handle any error when reading (i.e. user closed window)
		if err != nil {
			// logger.Debug("request failed")

			// unregister the client from the hub
			registration := newRegistration(client)
			hub.Unregister <- registration

			return nil
		}

		// parse the message content
		var msg Message
		err = json.Unmarshal(dta, &msg)
		if err != nil {
			logger.Debug("could not unmarshal request", logger.Err(err),
				logger.String("content", string(dta)))
		}

		msg.Raw = dta
		msg.DocumentID = client.DocumentID
		msg.UserID = client.UserID
		msg.Client = client
		msg.Permission = client.Permission

		// initialize the client handling and lazy load client permissions
		if msg.Type == MessageTypeProsemirrorInit {

			var payload ProsemirrorInitMessage
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				logger.DebugError("could not parse load message", err,
					logger.String("userid", client.UserID))
			}

			// register the client in the hub if it does not have a room associated yet
			if client.DocumentID == "" {

				// save the document id in the client to know which room the client
				// should belong to
				client.DocumentID = payload.DocumentID
				msg.DocumentID = payload.DocumentID
				client.DocumentSchema = payload.DocumentSchema

				p, err := hub.Srv.Postgres.FetchPermission(payload.DocumentID, client.UserID)
				if err != nil {
					logger.Debug("could not fetch permission for client", logger.String("userid", client.UserID),
						logger.String("documentid", client.DocumentID))
				}

				client.Permission = p
				msg.Permission = p

				if msg.Permission == domain.None {
					logger.Debug("permission denied. client not registered")
					return nil
				}

				// register the client in a document room
				registration := newRegistration(client)
				hub.Register <- registration
				<-registration.Done

				logger.Debug("client registered", logger.String("userid", client.UserID),
					logger.String("documentid", client.DocumentID),
					logger.String("permission", msg.Permission.String()))
			}

		}

		if msg.Permission == domain.None {
			logger.Debug("permission denied. message not handled")
			return nil
		}

		// store the client send channel as callback channel on the message
		msg.Reply = client.Send

		// send the message to the message handler of the room
		client.MessageHandler <- msg

	}

}
