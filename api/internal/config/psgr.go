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

// LoadPostgres loads Postgres databse configuration and returns printable error if at least one field is missing
func LoadPostgres() (*PostgresConfig, error) {
	conf := PostgresConfig{}
	err := &ErrLoadPostgresConf{}

	fillUpPostgresConf(
		[]*string{&conf.User, &conf.Password, &conf.DBName, &conf.DBHost},
		[]string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "DB_HOST"},
		err,
	)

	return &conf, err.Err()
}

// fillUpPostgresConf is a helper function to fill up configurration struct and collect all the missing fields
func fillUpPostgresConf(fields []*string, keys []string, err *ErrLoadPostgresConf) {
	for i, key := range keys {
		value, exists := os.LookupEnv(key)
		if exists {
			*fields[i] = value
		} else {
			err.AddField(key)
		}
	}
}
