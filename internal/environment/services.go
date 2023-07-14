package environment

import (
	"dkfbasel.ch/orca/collaboration/src/repository"
	image "dkfbasel.ch/orca/image/src/domain"
	process "dkfbasel.ch/orca/process/src/domain"
	"github.com/go-redis/redis/v7"
)

// Services is used to provide all necessary services
type Services struct {
	// redis database to store transactions in memory
	Redis *redis.Client

	// postgres database
	Postgres *repository.DB

	// process service to handle document
	Process process.ProcessClient

	// image service to handle images
	Image image.ImageClient
}
