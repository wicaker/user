package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/test/dbfixture"
)

func TestUserRepositoryFindAll(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	userRepo := repository.NewUserSqlxRepository(dbConn)

	// prepare data
	_, err := dbfixture.SeedUsers(dbConn, 5)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find all user", func(t *testing.T) {
		users, err := userRepo.FindAll(context.TODO())
		assert.NotEmpty(t, users)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
	})

	t.Run("no user registered", func(t *testing.T) {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
		users, err := userRepo.FindAll(context.TODO())
		assert.Empty(t, users)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})
}

func TestUserRepositoryFind(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	userRepo := repository.NewUserSqlxRepository(dbConn)

	// prepare data
	users, err := dbfixture.SeedUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find user", func(t *testing.T) {
		usrs, err := userRepo.Find(context.TODO(), users[0].UUID)

		assert.NotNil(t, usrs)
		assert.NoError(t, err)
		assert.Equal(t, users[0].UUID, usrs.UUID)
	})

	t.Run("no rows in result set", func(t *testing.T) {
		usrs, err := userRepo.Find(context.TODO(), uuid.New().String())
		assert.Nil(t, usrs)
		assert.NoError(t, err)
	})
}

func TestUserRepositoryFindOneBy(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	userRepo := repository.NewUserSqlxRepository(dbConn)

	// prepare data
	users, err := dbfixture.SeedUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find user", func(t *testing.T) {
		usrs, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": &users[0].Email,
		}, nil)

		assert.NotNil(t, usrs)
		assert.NoError(t, err)
		assert.Equal(t, users[0].UUID, usrs.UUID)
		assert.Equal(t, users[0].Email, usrs.Email)
	})

	t.Run("no rows in result set", func(t *testing.T) {
		email := "emailnotfound@example.com"
		usrs, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": &email,
		}, nil)
		assert.Nil(t, usrs)
		assert.NoError(t, err)
	})
}

func TestUserRepositoryStore(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	userRepo := repository.NewUserSqlxRepository(dbConn)

	// prepare data
	user := &domain.User{
		Email:    "user_repository_store@example.com",
		Password: "pass1",
	}

	// store new user
	t.Run("success store user", func(t *testing.T) {
		usr, err := userRepo.Store(context.TODO(), user)
		assert.NotNil(t, usr)
		assert.NoError(t, err)
		assert.Equal(t, time.Now().UTC().Format("2006-01-02 15:04"), usr.UpdatedAt.Format("2006-01-02 15:04"))
		assert.Equal(t, time.Now().UTC().Format("2006-01-02 15:04"), usr.CreatedAt.Format("2006-01-02 15:04"))
		assert.Equal(t, usr.CreatedAt, usr.UpdatedAt)
		assert.NotEmpty(t, usr.UUID)
	})

	t.Run("failed duplicate email", func(t *testing.T) {
		usr, err := userRepo.Store(context.TODO(), user)
		assert.Nil(t, usr)
		assert.Error(t, err)
	})

	t.Run("failed no email", func(t *testing.T) {
		usernoemail := &domain.User{
			Password: "pass1",
		}
		usr, err := userRepo.Store(context.TODO(), usernoemail)
		assert.Nil(t, usr)
		assert.Error(t, err)
	})

	t.Run("failed no password", func(t *testing.T) {
		usernopass := &domain.User{
			Email: "user_repository_store1@example.com",
		}
		usr, err := userRepo.Store(context.TODO(), usernopass)
		fmt.Println(err)
		assert.Nil(t, usr)
		assert.Error(t, err)
	})
}

func TestUserRepositoryUpdate(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	userRepo := repository.NewUserSqlxRepository(dbConn)

	// prepeare data
	users, err := dbfixture.SeedUsers(dbConn, 2)
	if err != nil {
		t.Error(err)
	}

	t.Run("success update user", func(t *testing.T) {
		user0 := users[0]
		user0.Email = "newEmail@example.com"

		updatedUser, err := userRepo.Update(context.TODO(), &user0)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)

		// get a user after updated (data appropriate)
		usr, err := userRepo.Find(context.TODO(), user0.UUID)
		assert.NotNil(t, usr)
		assert.NoError(t, err)
		assert.Equal(t, user0.Email, usr.Email)
		assert.Equal(t, user0.Password, usr.Password)
		assert.Equal(t, users[0].Password, usr.Password)
		assert.NotEqual(t, users[0].Email, usr.Email)
		assert.NotEqual(t, users[0].UpdatedAt, usr.UpdatedAt)
		assert.Equal(t, users[0].Password, usr.Password)
		assert.Equal(t, users[0].CreatedAt, usr.CreatedAt)
	})

	t.Run("failed duplicate email", func(t *testing.T) {
		user0 := users[0]
		user0.Email = users[1].Email
		updatedUser, err := userRepo.Update(context.TODO(), &user0)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
	})

	t.Run("failed no email", func(t *testing.T) {
		user0 := users[0]
		usernoemail := &domain.User{
			UUID:     user0.UUID,
			Password: "pass1",
		}
		updatedUser, err := userRepo.Update(context.TODO(), usernoemail)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
	})

	t.Run("failed no password", func(t *testing.T) {
		user0 := users[0]
		usernopass := &domain.User{
			UUID:  user0.UUID,
			Email: "user_repository_store1@example.com",
		}
		updatedUser, err := userRepo.Update(context.TODO(), usernopass)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
	})
}
