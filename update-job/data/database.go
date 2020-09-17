package data

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var data = struct {
	dsn string
	db  *gorm.DB
}{}

// SetDSN set DSN string for database connection
func SetDSN(dsn string) {
	data.dsn = dsn
}

// GetDB gets the db connection instance
func GetDB() (*gorm.DB, error) {
	if data.db == nil {
		if data.dsn == "" {
			return nil, errors.New("Cannot connect to database. DSN is not set")
		}
		db, err := gorm.Open(postgres.Open(data.dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		data.db = db
	}
	return data.db, nil
}
