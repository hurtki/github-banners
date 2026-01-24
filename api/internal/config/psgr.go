package config

import (
	"fmt"
	"os"
)

type PostgresConfig struct {
	User     string
	Password string
	DBName   string
	DBHost   string
	DBPort   string
}

type ErrLoadPostgresConf struct {
	fields []string
}

func (e ErrLoadPostgresConf) Error() string {
	return fmt.Sprintf("error, when loading postgres config, missing fields: %v", e.fields)
}

// Err turns out error interface, that is nil if ErrLoadPostgresConf contains no fields
func (e *ErrLoadPostgresConf) Err() error {
	if len(e.fields) == 0 {
		return nil
	} else {
		return e
	}
}

func (e *ErrLoadPostgresConf) AddField(field string) {
	e.fields = append(e.fields, field)
}

// LoadPostgres loads Postgres database configuration and returns printable error if at least one field is missing
func LoadPostgres() (*PostgresConfig, error) {
	conf := PostgresConfig{}
	err := &ErrLoadPostgresConf{}

	fillUpPostgresConf(
		map[string]*string{
			"POSTGRES_USER":     &conf.User,
			"POSTGRES_PASSWORD": &conf.Password,
			"POSTGRES_DB":       &conf.DBName,
			"DB_HOST":           &conf.DBHost,
			"PGPORT":            &conf.DBPort,
		},
		err,
	)

	return &conf, err.Err()
}

// fillUpPostgresConf is a helper function to fill up configuration struct and collect all the missing fields
func fillUpPostgresConf(vars map[string]*string, err *ErrLoadPostgresConf) {
	for envKey, fieldPtr := range vars {
		value, exists := os.LookupEnv(envKey)
		if exists {
			*fieldPtr = value
		} else {
			err.AddField(envKey)
		}
	}
}
