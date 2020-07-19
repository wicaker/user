package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/test/dbfixture"
)

func TestProfileRepositoryFindAll(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	_, err := dbfixture.SeedProfiles(dbConn, 5)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find all profile", func(t *testing.T) {
		profiles, err := profileRepo.FindAll(context.TODO())
		assert.NotEmpty(t, profiles)
		assert.NoError(t, err)
		assert.Len(t, profiles, 5)
		assert.Empty(t, profiles[0].Dob)
		assert.Empty(t, profiles[0].User)
	})

	t.Run("no profile registered", func(t *testing.T) {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
		profiles, err := profileRepo.FindAll(context.TODO())
		assert.Empty(t, profiles)
		assert.NoError(t, err)
		assert.Len(t, profiles, 0)
	})
}

func TestProfileRepositoryFindBy(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	profileFixtures, err := dbfixture.SeedProfiles(dbConn, 5)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find profiles by ... with limit", func(t *testing.T) {
		limit := uint(1)
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"dob": nil,
		}, nil, &limit, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, profiles)
		assert.Len(t, profiles, 1)
	})

	t.Run("success find profiles by ... with offset", func(t *testing.T) {
		offset := uint(1)
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"dob": nil,
		}, nil, nil, &offset)
		assert.NoError(t, err)
		assert.NotEmpty(t, profiles)
		assert.Len(t, profiles, 4)
	})

	t.Run("success find profiles by ... with limit and offset", func(t *testing.T) {
		limit := uint(1)
		offset := uint(1)
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"dob": nil,
		}, &map[string]string{
			"first_name": "ASC",
		}, &limit, &offset)
		assert.NoError(t, err)
		assert.NotEmpty(t, profiles)
		assert.Len(t, profiles, 1)
		assert.Equal(t, profileFixtures[1].FirstName, profiles[0].FirstName)
		assert.Equal(t, profileFixtures[1].UUID, profiles[0].UUID)
		assert.Equal(t, profileFixtures[1].UserUUID, profiles[0].UserUUID)
		assert.Empty(t, profiles[0].Dob)
	})

	t.Run("success find profiles by ...", func(t *testing.T) {
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, profiles)
		assert.Len(t, profiles, 1)
		assert.Equal(t, profileFixtures[0].FirstName, profiles[0].FirstName)
		assert.Equal(t, profileFixtures[0].UUID, profiles[0].UUID)
		assert.Equal(t, profileFixtures[0].UserUUID, profiles[0].UserUUID)
		assert.Empty(t, profiles[0].Dob)
		assert.Empty(t, profiles[0].User)
	})

	t.Run("success find profiles by ... with ORDER BY", func(t *testing.T) {
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"first_name": "ASC",
			"dob":        "DESC",
		}, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, profiles)
		assert.Len(t, profiles, 1)
		assert.Equal(t, profileFixtures[0].FirstName, profiles[0].FirstName)
		assert.Equal(t, profileFixtures[0].UUID, profiles[0].UUID)
		assert.Equal(t, profileFixtures[0].UserUUID, profiles[0].UserUUID)
		assert.Empty(t, profiles[0].Dob)
	})

	t.Run("failed find profiles because column does not exist", func(t *testing.T) {
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"random":     nil,
		}, nil, nil, nil)
		assert.Error(t, err)
		assert.Empty(t, profiles)
	})

	t.Run("failed find profiles because wrong orderBy input 1", func(t *testing.T) {
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"first_name": "ASCC",
			"dob":        "DESC",
		}, nil, nil)
		assert.Error(t, err)
		assert.Empty(t, profiles)
	})

	t.Run("failed find profiles because wrong orderBy input 2", func(t *testing.T) {
		profiles, err := profileRepo.FindBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"random": "DESC",
		}, nil, nil)
		assert.Error(t, err)
		assert.Empty(t, profiles)
	})
}

func TestProfileRepositoryFind(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	profileFixtures, err := dbfixture.SeedProfiles(dbConn, 5)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find profile", func(t *testing.T) {
		profile, err := profileRepo.Find(context.TODO(), profileFixtures[0].UUID)
		assert.NotEmpty(t, profile)
		assert.NoError(t, err)
	})

	t.Run("no profile found", func(t *testing.T) {
		profile, err := profileRepo.Find(context.TODO(), "8d47d418-83c6-4c00-ae82-d1aeb53c4fd2")
		assert.Empty(t, profile)
		assert.NoError(t, err)
	})

	t.Run("failed uuid", func(t *testing.T) {
		profile, err := profileRepo.Find(context.TODO(), "as79")
		assert.Empty(t, profile)
		assert.Error(t, err)
	})
}

func TestProfileRepositoryFindOneBy(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	profileFixtures, err := dbfixture.SeedProfiles(dbConn, 3)
	if err != nil {
		t.Error(err)
	}

	t.Run("success find a profile by ...", func(t *testing.T) {
		profile, err := profileRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, profile)
		assert.Equal(t, profileFixtures[0].FirstName, profile.FirstName)
		assert.Equal(t, profileFixtures[0].UUID, profile.UUID)
		assert.Equal(t, profileFixtures[0].UserUUID, profile.UserUUID)
		assert.Empty(t, profile.Dob)
		assert.Empty(t, profile.User)
	})

	t.Run("success find a profile by ... with ORDER BY", func(t *testing.T) {
		profile, err := profileRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"first_name": "ASC",
			"dob":        "DESC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, profile)
		assert.Equal(t, profileFixtures[0].FirstName, profile.FirstName)
		assert.Equal(t, profileFixtures[0].UUID, profile.UUID)
		assert.Equal(t, profileFixtures[0].UserUUID, profile.UserUUID)
		assert.Empty(t, profile.Dob)
	})

	t.Run("failed find a profile because column does not exist", func(t *testing.T) {
		profile, err := profileRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"random":     nil,
		}, nil)
		assert.Error(t, err)
		assert.Empty(t, profile)
	})

	t.Run("failed find a profile because wrong orderBy input 1", func(t *testing.T) {
		profile, err := profileRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"first_name": "ASCC",
			"dob":        "DESC",
		})
		assert.Error(t, err)
		assert.Empty(t, profile)
	})

	t.Run("failed find a profile because wrong orderBy input 2", func(t *testing.T) {
		profile, err := profileRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"first_name": profileFixtures[0].FirstName,
			"dob":        nil,
			"user_uuid":  &profileFixtures[0].UserUUID,
		}, &map[string]string{
			"random": "DESC",
		})
		assert.Error(t, err)
		assert.Empty(t, profile)
	})
}

func TestProfileRepositoryStore(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	users, err := dbfixture.SeedUsers(dbConn, 3)
	if err != nil {
		t.Error(err)
	}

	firstName := fmt.Sprintf("FirstName%d", 1)
	lastName := fmt.Sprintf("LastName%d", 1)
	phone := fmt.Sprintf("628191919%d", 1)
	address := fmt.Sprintf("Street %d", 1)

	// store new profile
	t.Run("success store a profile", func(t *testing.T) {
		profile := domain.Profile{
			User:      users[0],
			FirstName: &firstName,
			LastName:  &lastName,
			Phone:     &phone,
			Address:   &address,
		}

		profileResult, err := profileRepo.Store(context.TODO(), &profile)
		assert.NotNil(t, profileResult)
		assert.NoError(t, err)
		assert.Equal(t, time.Now().UTC().Format("2006-01-02 15:04"), profileResult.UpdatedAt.Format("2006-01-02 15:04"))
		assert.Equal(t, time.Now().UTC().Format("2006-01-02 15:04"), profileResult.CreatedAt.Format("2006-01-02 15:04"))
		assert.Equal(t, profileResult.CreatedAt, profileResult.UpdatedAt)
		assert.Equal(t, profile.User, profileResult.User)
		assert.NotEmpty(t, profileResult.UUID)
	})

	t.Run("failed duplicate user", func(t *testing.T) {
		profile := domain.Profile{
			User:      users[0],
			FirstName: &firstName,
			LastName:  &lastName,
			Phone:     &phone,
			Address:   &address,
		}

		profileResult, err := profileRepo.Store(context.TODO(), &profile)
		assert.Nil(t, profileResult)
		assert.Error(t, err)
	})

	t.Run("failed gender input", func(t *testing.T) {
		gender := "random"
		profile := domain.Profile{
			User:      users[0],
			FirstName: &firstName,
			LastName:  &lastName,
			Phone:     &phone,
			Address:   &address,
			Gender:    &gender,
		}

		profileResult, err := profileRepo.Store(context.TODO(), &profile)
		assert.Nil(t, profileResult)
		assert.Error(t, err)
	})
}

func TestProfileRepositoryUpdate(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()
	profileRepo := repository.NewProfileSqlxRepository(dbConn)

	// prepare data
	profileFixtures, err := dbfixture.SeedProfiles(dbConn, 2)
	if err != nil {
		t.Error(err)
	}

	// update a profile
	t.Run("success update a profile", func(t *testing.T) {
		firstName := fmt.Sprintf("FirstNameUpdated%d", 1)
		lastName := fmt.Sprintf("LastNameUpdated%d", 1)

		profile := profileFixtures[0]
		//before update
		assert.Equal(t, "FirstName0", *profile.FirstName)
		assert.Equal(t, "LastName0", *profile.LastName)

		profile.FirstName = &firstName
		profile.LastName = &lastName
		err := profileRepo.Update(context.TODO(), &profile)
		assert.NoError(t, err)

		// after updated (data appropriate)
		profileResult, err := profileRepo.Find(context.TODO(), profile.UUID)
		assert.NotNil(t, profileResult)
		assert.NoError(t, err)
		assert.Equal(t, profile.FirstName, profileResult.FirstName)
		assert.Equal(t, profile.LastName, profileResult.LastName)
		assert.Equal(t, profile.UserUUID, profileResult.UserUUID)
	})

	t.Run("failed duplicate user", func(t *testing.T) {
		firstName := fmt.Sprintf("FirstNameUpdated%d", 1)
		lastName := fmt.Sprintf("LastNameUpdated%d", 1)

		profile := profileFixtures[0]
		profile.FirstName = &firstName
		profile.LastName = &lastName
		profile.User = profileFixtures[1].User

		err := profileRepo.Update(context.TODO(), &profile)
		assert.Error(t, err)
	})

	t.Run("failed gender input", func(t *testing.T) {
		gender := "random"
		profile := profileFixtures[0]
		profile.Gender = &gender

		err := profileRepo.Update(context.TODO(), &profile)
		assert.Error(t, err)
	})
}
