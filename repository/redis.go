package repository

import (
	"github.com/go-redis/redis/v7"
)

// NewRedisClient will initialize an connection to the redis server
func NewRedisClient(host, password string) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0, // use default database
	})

	return client, nil

}
