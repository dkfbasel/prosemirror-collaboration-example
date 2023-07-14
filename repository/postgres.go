package repository

import (
	"dkfbasel.ch/orca/pkg/database"
	"dkfbasel.ch/orca/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// DB will implement the method to satisfy the sampleDBInterface
type DB struct {
	Session *sqlx.DB
}

// NewPostgresClient will initialize a connection to the postgres database
func NewPostgresClient(config database.Config) (*DB, error) {

	session, err := database.NewSession(config)
	if err != nil {
		logger.Info("could not open database session")
	}

	db := DB{}
	db.Session = session

	return &db, nil
}

// Close will terminate the database connections
func (db *DB) Close() error {
	return db.Session.Close()
}
