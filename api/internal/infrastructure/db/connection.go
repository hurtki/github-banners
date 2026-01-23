package db

import (
	"database/sql"

	"fmt"
	"time"

	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// DB_MAX_CONNECTIONS_PERCENT_TO_POSTGRES_MAX - percent of overall db connections from postgres max connections
	dbMaxConnectionsPercentToPostgresMax = 70
	// DB_MAX_IDLE_CONNECTIONS_PERCENT_TO_POSTGRES_MAX - percent of db idling connections from postgres max connections
	dbMaxIdleConnectionsPercentToPostgresMax = 10

	connectionLifeTime           = 5 * time.Minute
	dataBaseConnectionTriesCount = 10
	retryTimeBetweenTries        = 1 * time.Second
)

func NewDB(conf *config.PostgresConfig, logger logger.Logger) (*sql.DB, error) {
	fn := "internal.infrastructure.db.NewDB"

	logger = logger.With("service", "db-intialization-function")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", conf.User, conf.Password, conf.DBHost, conf.DBPort, conf.DBName)

	var db *sql.DB
	var err error

	for i := range dataBaseConnectionTriesCount {
		// Connection with data base
		db, err = sql.Open("pgx", dsn)

		if err != nil {
			logger.Warn(fmt.Sprintf("Cannot open database(retrying in %s), try number: %d/%d", retryTimeBetweenTries.String(), i+1, dataBaseConnectionTriesCount), "source", fn)
			time.Sleep(retryTimeBetweenTries)
			continue
		}

		if err := db.Ping(); err != nil {
			logger.Warn(fmt.Sprintf("Cannot ping database, try number: %d/%d", i+1, dataBaseConnectionTriesCount), "source", fn, "err", err)
			db.Close()
			time.Sleep(retryTimeBetweenTries)
			continue
		}
		logger.Info("esablished connection with database")
		break
	}

	if err != nil {
		logger.Error("Can't establish connection with database", "source", fn, "err", err)
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	maxConnectionsRow := db.QueryRow(`
	SHOW max_connections;
	`)
	var maxPostgresConnections int
	if err := maxConnectionsRow.Scan(&maxPostgresConnections); err != nil {
		logger.Error("cannot get postgres max connections", "err", err, "source", fn)
		return nil, fmt.Errorf("can't get postgres max connections, can't intialize storage: %w", err)
	}

	logger.Info("got postgres max connections count", "count", maxPostgresConnections, "source", fn)

	var dbMaxOpenCons int = maxPostgresConnections * dbMaxConnectionsPercentToPostgresMax / 100
	var dbMaxIdleCons int = maxPostgresConnections * dbMaxIdleConnectionsPercentToPostgresMax / 100

	logger.Info("setting db max open connections", "count", dbMaxOpenCons, "source", fn)
	logger.Info("setting db max idle connections", "count", dbMaxIdleCons, "source", fn)

	db.SetMaxOpenConns(dbMaxOpenCons)
	db.SetMaxIdleConns(dbMaxIdleCons)
	logger.Info("setting connection life time", "count", connectionLifeTime.String(), "source", fn)
	db.SetConnMaxLifetime(connectionLifeTime)
	return db, nil
}
