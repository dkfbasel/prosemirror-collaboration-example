package websocket

import (
	domain "dkfbasel.ch/orca/collaboration/src/domain"
	"dkfbasel.ch/orca/collaboration/src/internal/environment"
	"dkfbasel.ch/orca/pkg/logger"
)

// handleMessage will triage incoming messages
func handleMessage(srv *environment.Services, room *WebsocketRoom,
	message *Message) {

	if message.Permission != domain.Edit {
		logger.Debug("permission denied")
		return
	}

	switch message.Type {

	case MessageTypeProsemirrorInit:
		logger.Debug("handle prosemirror init")
		handleProsemirrorInitMessage(srv, room, message)
		handleProsemirrorStepsMessage(srv, room, message, true)

	case MessageTypeProsemirrorUpdate:
		logger.Debug("handle prosemirror update")
		handleProsemirrorStepsMessage(srv, room, message, false)

	case MessageTypeProsemirrorSteps:
		logger.Debug("handle prosemirror steps")
		handleProsemirrorStepsMessage(srv, room, message, false)
	}

}
