package config

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"

	// import source driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// SqlxConfig sqlx configuration
type SqlxConfig struct {
	url string
	mi  *migrate.Migrate
}

// NewSqlx will create new a SqlxConfig represent configuration database using sqlx library
func NewSqlx() *SqlxConfig {
	config := new(SqlxConfig)

	config.url = os.Getenv("DATABASE_URL")
	return config
}

// Open will open database connection using sqlx library
func (s *SqlxConfig) Open() (*sqlx.DB, error) {
	log.Println("open database connection...")
	return sqlx.Open(`postgres`, s.url)
}

// Close database connection
func (s *SqlxConfig) Close(conn *sqlx.DB) error {
	log.Println("closing database connection...")
	return conn.Close()
}

// MigrateUp running migration up
func (s *SqlxConfig) MigrateUp(filePath string) error {
	mi, err := migrate.New(
		filePath,
		s.url)

	if err != nil {
		return err
	}

	s.mi = mi

	log.Println("migrating up...")
	return mi.Up()
}

// MigrateDown running migration
func (s *SqlxConfig) MigrateDown() error {
	log.Println("migrating down...")
	return s.mi.Down()
}
