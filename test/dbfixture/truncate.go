package dbfixture

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Truncate table
func Truncate(dbConn *sqlx.DB) error {
	stmt := "TRUNCATE TABLE users, profiles;"

	if _, err := dbConn.Exec(stmt); err != nil {
		return errors.Wrap(err, "truncate test database tables")
	}

	return nil
}
