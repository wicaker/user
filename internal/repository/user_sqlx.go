package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/wicaker/user/internal/domain"
)

type userSqlxRepository struct {
	conn *sqlx.DB
}

// NewUserSqlxRepository will create new an userSqlxRepository object representation of domain.UserRepository interface
func NewUserSqlxRepository(conn *sqlx.DB) domain.UserRepository {
	return &userSqlxRepository{conn}
}

func (db *userSqlxRepository) Find(ctx context.Context, uuid string) (*domain.User, error) {
	user := new(domain.User)
	err := db.conn.GetContext(ctx, user, `SELECT * FROM users WHERE uuid=$1`, uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (db *userSqlxRepository) FindOneBy(ctx context.Context, criteria map[string]interface{}, orderBy *map[string]string) (*domain.User, error) {
	var (
		user              = new(domain.User)
		filterQuery, args = filterRecordsQuery(criteria, orderBy)
	)

	err := db.conn.GetContext(ctx, user, `SELECT * FROM users WHERE 1=1`+filterQuery, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (db *userSqlxRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User

	err := db.conn.SelectContext(ctx, &users, `SELECT * FROM users`)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (db *userSqlxRepository) FindBy(ctx context.Context, criterias map[string]interface{}, orderBy *map[string]string, limit *uint, offset *uint) ([]*domain.User, error) {
	var (
		users             []*domain.User
		filterQuery, args = filterRecordsQuery(criterias, orderBy)
		offsetAndLimit    string
	)

	if nil != limit {
		offsetAndLimit = offsetAndLimit + fmt.Sprintf(" LIMIT %d", *limit)
	}

	if nil != offset {
		offsetAndLimit = offsetAndLimit + fmt.Sprintf(" OFFSET %d", *offset)
	}

	err := db.conn.SelectContext(ctx, &users, `SELECT * FROM users WHERE 1=1`+filterQuery+offsetAndLimit, args...)
	if err != nil {
		return users, err
	}
	return users, nil
}

func (db *userSqlxRepository) Store(ctx context.Context, user *domain.User) (*domain.User, error) {
	stmt, err := db.conn.PrepareContext(ctx, "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING uuid, salt, created_at, updated_at")
	if err != nil {
		return nil, errors.Wrap(err, "prepare users insertion")
	}

	row := stmt.QueryRow(user.Email, user.Password)

	if err = row.Scan(&user.UUID, &user.Salt, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}

		return nil, errors.Wrap(err, "row scan")
	}

	if err := stmt.Close(); err != nil {
		return nil, errors.Wrap(err, "close psql statement")
	}

	return user, err
}

func (db *userSqlxRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	stmt, err := db.conn.PrepareContext(ctx, `UPDATE users SET email=$1 , password=$2, is_active=$3, new_password=$4 WHERE uuid=$5 RETURNING uuid, salt, created_at, updated_at`)
	if err != nil {
		return nil, errors.Wrap(err, "prepare users update")
	}

	row := stmt.QueryRow(
		user.Email,
		user.Password,
		user.IsActive,
		user.NewPassword,
		user.UUID,
	)

	if err = row.Scan(&user.UUID, &user.Salt, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err := stmt.Close(); err != nil {
			return nil, errors.Wrap(err, "close psql statement")
		}

		return nil, errors.Wrap(err, "row scan")
	}

	if err := stmt.Close(); err != nil {
		return nil, errors.Wrap(err, "close psql statement")
	}

	return user, err
}
