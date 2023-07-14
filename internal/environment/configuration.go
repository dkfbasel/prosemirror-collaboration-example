package environment

import (
	"dkfbasel.ch/orca/pkg/database"
	"github.com/kelseyhightower/envconfig"
)

// Configuration holds the basic config information for the service
type Configuration struct {

	// host to start the httpserver on
	Websocket struct {
		Host string `default:"0.0.0.0:80"`
	}

	// redis server connection
	Redis struct {
		Host     string `default:"orca.redis:6379"`
		Password string `default:""`
	}

	// information about process service
	Process struct {
		Address string `default:"service.process"`
	}

	// information about image service
	Image struct {
		Address string `default:"service.image"`
	}

	// database configuration for the postgres connection
	Postgres database.Config `envconfig:"DB"`
}

// LoadConfiguration will load the basic application configuration from the
// specified config file
func LoadConfiguration(prefix string) (Configuration, error) {
	config := Configuration{}
	err := envconfig.Process(prefix, &config)
	return config, err
}
