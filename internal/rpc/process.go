package rpc

import (
	"dkfbasel.ch/orca/pkg/ctxvalue"
	"dkfbasel.ch/orca/pkg/logger"
	process "dkfbasel.ch/orca/process/src/domain"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// NewProcessClient initializes a grpc connection to the process service
func NewProcessClient(address string) (process.ProcessClient, error) {

	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			logger.Grpc(),
			ctxvalue.SetContext(),
		)))

	if err != nil {
		return nil, err
	}

	c := process.NewProcessClient(conn)
	return c, nil
}
