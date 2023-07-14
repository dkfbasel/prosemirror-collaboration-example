package websocket

import (
	"context"
	"time"

	"dkfbasel.ch/orca/pkg/logger"
	"nhooyr.io/websocket"
)

// handleSend will initialize a go routine to send information to the client
func handleSend(client *WebsocketClient) {

	// read all messages sent on the send channel and return it to the client
	for message := range client.Send {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		err := client.Conn.Write(ctx, websocket.MessageText, message)
		cancel()
		if err != nil {
			logger.Debug("could not write message")
		}
	}

}
