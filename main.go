package main

import (
	"net"
	"net/http"

	"dkfbasel.ch/orca/collaboration/src/internal/environment"
	"dkfbasel.ch/orca/collaboration/src/internal/rpc"
	"dkfbasel.ch/orca/collaboration/src/repository"
	"dkfbasel.ch/orca/collaboration/src/websocket"
	"dkfbasel.ch/orca/pkg/logger"
)

func main() {

	// ensure thata the logger is flushed when the application is terminated
	defer logger.Sync()

	// load the application configuration
	config, err := environment.LoadConfiguration("profiles")

	if err != nil {
		logger.FatalError("startup aborted, incomplete configuration", err)
	}

	// initialize service struct
	srv := environment.Services{}

	// establish a connection to the redis instance
	srv.Redis, err = repository.NewRedisClient(config.Redis.Host, config.Redis.Password)
	if err != nil {
		logger.FatalError("startup aborted. could not initialize redis connection", err)
	}
	defer srv.Redis.Close()

	// establish a new postgres connection
	srv.Postgres, err = repository.NewPostgresClient(config.Postgres)
	if err != nil {
		logger.FatalError("startup aborted. could not initialize postgres connection", err)
	}
	defer srv.Postgres.Close()

	// initialize connection to the process service
	srv.Process, err = rpc.NewProcessClient(config.Process.Address)
	if err != nil {
		logger.FatalError("startup aborted. could not initialize process service", err)
	}

	// initialize connection to the image service
	srv.Image, err = rpc.NewImageClient(config.Image.Address)
	if err != nil {
		logger.FatalError("startup aborted. could not initialize image service", err)
	}

	// start a tcp listener on the given port
	listener, err := net.Listen("tcp", config.Websocket.Host)
	if err != nil {
		logger.FatalError("failed to start tcp server", err)
	}
	defer listener.Close()

	// define the websocket server
	wsServer := websocket.NewServer(&srv)
	defer wsServer.Close() // nolint:errcheck

	logger.Info("starting server")

	// start the websocket server
	err = wsServer.Serve(listener)
	if err != http.ErrServerClosed {
		logger.FatalError("failed to listen and serve", err)
	}

}
