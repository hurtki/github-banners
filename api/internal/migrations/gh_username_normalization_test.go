package migrations

import (
	"log"
	"testing"

	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/infrastructure/db"
	"github.com/hurtki/github-banners/api/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func Test004Success(t *testing.T) {

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(t.Context(),
		"postgres:15",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)

	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	host, err := postgresContainer.Host(t.Context())
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(t.Context(), "5432")
	require.NoError(t, err)

	cfg := config.PostgresConfig{User: dbUser, DBName: dbName, Password: dbPassword, DBHost: host, DBPort: port.Port()}

	logger := logger.NewLogger("ERROR", "TEXT")

	db, err := db.NewDB(&cfg, logger)
	require.NoError(t, err)

	err = goose.UpTo(db, ".", 3)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(), `INSERT INTO users (
    username,
    name,
    company,
    location,
    bio,
    public_repos_count,
    followers_count,
    following_count,
    fetched_at
) VALUES
('hurtki', 'Ivan Petrov', 'Example Corp', 'Haifa, Israel', 'Backend developer working with Go', 42, 120, 75, NOW()),
('HURTKI', 'Ivan Petrov', 'Example Corp', 'Haifa, Israel', 'Uppercase username test', 42, 120, 75, NOW()),
('HurtKi', 'Ivan Petrov', 'Example Corp', 'Haifa, Israel', 'Mixed case username test', 42, 120, 75, NOW()),
('johnDoe', 'John Doe', 'Acme Inc', 'New York, USA', 'Software engineer', 15, 80, 20, NOW()),
('JOHNDOE', 'John Doe', 'Acme Inc', 'New York, USA', 'Uppercase duplicate test', 15, 80, 20, NOW()),
('JaneSmith', 'Jane Smith', 'TechSoft', 'London, UK', 'Full-stack developer', 27, 150, 60, NOW()),
('janesmith', 'Jane Smith', 'TechSoft', 'London, UK', 'Lowercase duplicate test', 27, 150, 60, NOW()),
('DEV_GUY', 'Alex Brown', 'Startup Labs', 'Berlin, Germany', 'Open source contributor', 9, 40, 10, NOW()),
('dev_guy', 'Alex Brown', 'Startup Labs', 'Berlin, Germany', 'Case variant with underscore', 9, 40, 10, NOW()),
('SomeUser123', 'Chris White', 'DataWorks', 'Toronto, Canada', 'Data engineer', 33, 95, 44, NOW()),
('someuser123', 'Chris White', 'DataWorks', 'Toronto, Canada', 'Case duplicate with numbers', 33, 95, 44, NOW()),
('MiXeDCaSeUser', 'Taylor Green', 'CloudNine', 'Sydney, Australia', 'Cloud infrastructure engineer', 21, 60, 18, NOW());`)
	require.NoError(t, err)

	err = goose.UpTo(db, ".", 4)
	require.NoError(t, err)

	rows, err := db.QueryContext(t.Context(), `
    SELECT username_normalized
    FROM github_data.users
    WHERE username_normalized != lower(username);
`)
	require.NoError(t, err)
	defer rows.Close()

	var invalid []string
	for rows.Next() {
		var u string
		err := rows.Scan(&u)
		require.NoError(t, err)
		invalid = append(invalid, u)
	}

	require.NoError(t, rows.Err())
	require.Empty(t, invalid, "all usernames must be normalized to lowercase")
}
