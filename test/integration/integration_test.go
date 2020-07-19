package integration_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate"
	"github.com/jmoiron/sqlx"

	"github.com/wicaker/user/config"
	"github.com/wicaker/user/test/dbfixture"
)

var (
	dbConn       *sqlx.DB
	migrate_down string
)

func init() {
	os.Setenv("DATABASE_URL", "postgres://root:root@localhost:5432/microservice_user?sslmode=disable")
	os.Setenv("RABBITMQ_SERVER", "amqp://guest:guest@localhost:5672/")
	migrate_down = os.Getenv("migrate_down")
}

func TestMain(m *testing.M) {
	var (
		err      error
		sqlxConf = config.NewSqlx()
	)

	// open db connection
	dbConn, err = sqlxConf.Open()
	if err != nil {
		log.Fatal(err)
	}

	// migrate up
	err = sqlxConf.MigrateUp("file://../../migrations")
	if err != nil {
		if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", migrate.ErrNoChange) {
			log.Fatal(err)
		} else {
			if err := dbfixture.Truncate(dbConn); err != nil {
				log.Fatal(err)
			}
		}
	}

	// runs the tests
	code := m.Run()

	// migrate down
	if migrate_down == "true" {
		err = sqlxConf.MigrateDown()
		if err != nil {
			log.Fatal(err)
		}
	}

	// close db connection
	err = sqlxConf.Close(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
