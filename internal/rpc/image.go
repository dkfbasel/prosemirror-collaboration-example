package rpc

import (
	image "dkfbasel.ch/orca/image/src/domain"
	"dkfbasel.ch/orca/pkg/ctxvalue"
	"dkfbasel.ch/orca/pkg/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// NewImageClient initializes a grpc connection to the process service
func NewImageClient(address string) (image.ImageClient, error) {

	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			logger.Grpc(),
			ctxvalue.SetContext(),
		)))

	if err != nil {
		return nil, err
	}

	c := image.NewImageClient(conn)
	return c, nil
}
