package websocket

import (
	"encoding/json"

	domain "dkfbasel.ch/orca/collaboration/src/domain"
)

// Message types to differentiate the respective actions to take
type MessageType string

const MessageTypeProsemirrorInit MessageType = "prosemirror-init"
const MessageTypeProsemirrorUpdate MessageType = "prosemirror-update"
const MessageTypeProsemirrorSteps MessageType = "prosemirror-steps"
const MessageTypeProsemirrorApproval MessageType = "prosemirror-approval"
const MessageTypeProssemirrorReload MessageType = "prosemirror-reload"

type Message struct {
	Type    MessageType     `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload"`
	Raw     []byte          `json:"--"`

	// internal information
	DocumentID string            `json:"-"` // current document id
	UserID     string            `json:"-"` // current connected user
	Permission domain.Permission `json:"-"` // permission of the client

	// channel to reply to the sender
	Client *WebsocketClient `json:"-"`
	Reply  chan []byte      `json:"-"`
}

// Response sent back to the client
type Response struct {
	Type    MessageType `json:"type,omitempty"`
	Payload interface{} `json:"payload"`
}

// Encode will encode the current response for transfer
func (r *Response) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// NotifyMessage is used to send a message to all other
// clients except the sender
type NotifyMessage struct {
	Client  *WebsocketClient
	Payload json.RawMessage
}
