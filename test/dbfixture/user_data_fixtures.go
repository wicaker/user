package dbfixture

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/wicaker/user/internal/domain"
)

// SeedUsers handles seeding the users table in the database for integration tests
func SeedUsers(dbConn *sqlx.DB, count int) ([]domain.User, error) {
	var users []domain.User

	for i := 1; i <= count; i++ {
		user := domain.User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: fmt.Sprintf("Password%d", i),
		}

		// hash password
		password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		user.Password = string(password)

		stmt, err := dbConn.Prepare("INSERT INTO users (email, password) VALUES ($1, $2) RETURNING uuid, created_at, updated_at")
		if err != nil {
			return nil, errors.Wrap(err, "prepare users insertion")
		}

		row := stmt.QueryRow(user.Email, user.Password)

		if err = row.Scan(&user.UUID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			if err := stmt.Close(); err != nil {
				return nil, errors.Wrap(err, "close psql statement")
			}

			return nil, errors.Wrap(err, "capture users id")
		}

		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}

		users = append(users, user)
	}

	return users, nil
}

// SeedActiveUsers handles seeding the users table in the database for integration tests
func SeedActiveUsers(dbConn *sqlx.DB, count int) ([]domain.User, error) {
	var users []domain.User

	for i := 1; i <= count; i++ {
		user := domain.User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: fmt.Sprintf("Password%d", i),
		}

		// hash password
		password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		user.Password = string(password)

		stmt, err := dbConn.Prepare("INSERT INTO users (email, password, is_active) VALUES ($1, $2, $3) RETURNING uuid, created_at, updated_at")
		if err != nil {
			return nil, errors.Wrap(err, "prepare users insertion")
		}

		row := stmt.QueryRow(user.Email, user.Password, true)

		if err = row.Scan(&user.UUID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			if err := stmt.Close(); err != nil {
				return nil, errors.Wrap(err, "close psql statement")
			}

			return nil, errors.Wrap(err, "capture users id")
		}

		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}

		users = append(users, user)
	}

	return users, nil
}
