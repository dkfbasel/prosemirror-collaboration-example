package websocket

import (
	"net/http"
	"time"

	"dkfbasel.ch/orca/collaboration/src/internal/environment"
	"dkfbasel.ch/orca/pkg/logger"
)

// NewServer will initialize a new server to handle websocket connections
func NewServer(srv *environment.Services) *http.Server {

	// initialize a hub for connections
	hub := newHub(srv)

	return &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := wsHandler(hub, w, r)
			if err != nil {
				logger.DebugError("websocket handler,", err)
			}
		}),
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
	}
}
