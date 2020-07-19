package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/wicaker/user/internal/domain"
)

type profileSqlxRepository struct {
	conn *sqlx.DB
}

// NewProfileSqlxRepository will create new an profileSqlxRepository object representation of domain.ProfileRepository interface
func NewProfileSqlxRepository(conn *sqlx.DB) domain.ProfileRepository {
	return &profileSqlxRepository{conn}
}

func (db *profileSqlxRepository) Find(ctx context.Context, uuid string) (*domain.Profile, error) {
	profile := new(domain.Profile)
	err := db.conn.GetContext(ctx, profile, `SELECT * FROM profiles WHERE uuid=$1`, uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return profile, nil
}

func (db *profileSqlxRepository) FindOneBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string) (*domain.Profile, error) {
	var (
		profile           = new(domain.Profile)
		filterQuery, args = filterRecordsQuery(criteria, orderBy)
	)

	err := db.conn.GetContext(ctx, profile, `SELECT * FROM profiles WHERE 1=1`+filterQuery, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return profile, nil
}

func (db *profileSqlxRepository) FindAll(ctx context.Context) ([]*domain.Profile, error) {
	var profiles []*domain.Profile

	err := db.conn.SelectContext(ctx, &profiles, `SELECT * FROM profiles`)
	if err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (db *profileSqlxRepository) FindBy(ctx context.Context, criterias map[string]interface{}, orderBy *map[string]string, limit *uint, offset *uint) ([]*domain.Profile, error) {
	var (
		profiles          []*domain.Profile
		filterQuery, args = filterRecordsQuery(criterias, orderBy)
		offsetAndLimit    string
	)

	if nil != limit {
		offsetAndLimit = offsetAndLimit + fmt.Sprintf(" LIMIT %d", *limit)
	}

	if nil != offset {
		offsetAndLimit = offsetAndLimit + fmt.Sprintf(" OFFSET %d", *offset)
	}

	err := db.conn.SelectContext(ctx, &profiles, `SELECT * FROM profiles WHERE 1=1`+filterQuery+offsetAndLimit, args...)
	if err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (db *profileSqlxRepository) Store(ctx context.Context, profile *domain.Profile) (*domain.Profile, error) {
	stmt, err := db.conn.Prepare("INSERT INTO profiles (user_uuid, first_name, last_name, address, phone, gender, dob) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING uuid, user_uuid, created_at, updated_at")
	if err != nil {
		return nil, errors.Wrap(err, "prepare profiles insertion")
	}

	row := stmt.QueryRow(profile.User.UUID, profile.FirstName, profile.LastName, profile.Address, profile.Phone, profile.Gender, profile.Dob)

	if err = row.Scan(&profile.UUID, &profile.UserUUID, &profile.CreatedAt, &profile.UpdatedAt); err != nil {
		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}

		return nil, errors.Wrap(err, "row scan")
	}

	if err := stmt.Close(); err != nil {
		return nil, errors.Wrap(err, "close psql statement")
	}
	return profile, err

}

func (db *profileSqlxRepository) Update(ctx context.Context, profile *domain.Profile) error {
	stmt, err := db.conn.PrepareContext(ctx, `UPDATE profiles SET first_name=$1, last_name=$2, address=$3, phone=$4, gender=$5, dob=$6, user_uuid=$7 WHERE uuid=$8`)
	if err != nil {
		return errors.Wrap(err, "prepare profiles update")
	}

	_, err = stmt.ExecContext(
		ctx,
		profile.FirstName,
		profile.LastName,
		profile.Address,
		profile.Phone,
		profile.Gender,
		profile.Dob,
		profile.User.UUID,
		profile.UUID,
	)

	if err != nil {
		return errors.Wrap(err, "executes a update query")
	}

	if err := stmt.Close(); err != nil {
		return errors.Wrap(err, "close psql statement")
	}

	return err
}
