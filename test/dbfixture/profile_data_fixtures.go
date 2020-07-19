package dbfixture

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/wicaker/user/internal/domain"
)

// SeedProfiles handles seeding the profiles table in the database for integration tests
func SeedProfiles(dbConn *sqlx.DB, count int) ([]domain.Profile, error) {
	users, err := SeedUsers(dbConn, count)
	if err != nil {
		return nil, err
	}

	profiles := []domain.Profile{}
	for i, user := range users {
		firstName := fmt.Sprintf("FirstName%d", i)
		lastName := fmt.Sprintf("LastName%d", i)
		phone := fmt.Sprintf("628191919%d", i)
		address := fmt.Sprintf("Street %d", i)

		profiles = append(profiles, domain.Profile{
			User:      user,
			FirstName: &firstName,
			LastName:  &lastName,
			Phone:     &phone,
			Address:   &address,
		})
	}

	for i := range profiles {
		stmt, err := dbConn.Prepare("INSERT INTO profiles (user_uuid, first_name, last_name, address, phone, gender, dob) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING uuid, user_uuid, created_at, updated_at")
		if err != nil {
			return nil, errors.Wrap(err, "prepare profiles insertion")
		}

		row := stmt.QueryRow(profiles[i].User.UUID, profiles[i].FirstName, profiles[i].LastName, profiles[i].Address, profiles[i].Phone, profiles[i].Gender, profiles[i].Dob)

		if err = row.Scan(&profiles[i].UUID, &profiles[i].UserUUID, &profiles[i].CreatedAt, &profiles[i].UpdatedAt); err != nil {
			if err := stmt.Close(); err != nil {
				return nil, errors.Wrap(err, "close psql statement")
			}

			return nil, errors.Wrap(err, "capture users id")
		}

		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}
	}

	return profiles, nil
}
