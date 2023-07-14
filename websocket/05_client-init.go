package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"dkfbasel.ch/orca/collaboration/src/internal/session"
	"dkfbasel.ch/orca/pkg/logger"
	"nhooyr.io/websocket"
)

// WebsocketClient information
type WebsocketClient struct {
	// reference the websocket connection
	Conn *websocket.Conn

	// document is the unique id of the block that the client is working on
	DocumentID     string
	DocumentSchema json.RawMessage

	// unique id of the respective user
	UserID string

	// Document permissions of the respective client (read, comment, edit)
	Permission domain.Permission

	// use a channel to send messages
	Send chan []byte

	// reference to the handler that will manage the message
	MessageHandler chan Message
}

// Registration with separate done channel to wait until registration is complete
type Registration struct {
	Client *WebsocketClient
	Done   chan bool
}

// newRegistration will return a new registration for the given client
func newRegistration(client *WebsocketClient) *Registration {
	return &Registration{
		Client: client,
		Done:   make(chan bool),
	}
}

// wsHandler defines how to handle websocket requests
func wsHandler(hub *WebsocketHub, w http.ResponseWriter, r *http.Request) error {

	// parse the account id from the session header (passed by the auth service)
	sessionInfo, err := session.Parse(r.Header.Get("Session"))
	if err != nil {
		return logger.NewError("session information missing", err)
	}

	// initialize a new websocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// note: use insecureSkipVerify:true to allow access from other domains
		InsecureSkipVerify: false,
	})
	if err != nil {
		return fmt.Errorf("could not accept websocket connection: %w", err)
	}
	// set read limit to 3MB to avoid errors on pasting long text passages
	conn.SetReadLimit(3000000)
	defer conn.Close(websocket.StatusInternalError, "connection could not be established")

	// initialize a new websocket client
	client := &WebsocketClient{}
	client.UserID = sessionInfo.UserID
	client.Conn = conn
	client.Send = make(chan []byte)

	// initialize client with no permissions
	client.Permission = domain.None

	// initialize separate routine to send messages back to the client
	go handleSend(client)

	// handle incoming requests
	return handleReceive(hub, client)
}
