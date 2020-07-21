package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/test/dbfixture"
)

func TestForgotPasswordConfirmParsedTokenDataInvalid(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	users, err := dbfixture.SeedUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	var (
		userRepo    = repository.NewUserSqlxRepository(dbConn)
		newPassword = "newpassword"
		jwt         = createJWT(users[0], time.Minute*2)
		resp        domain.Response
		reqBody     = domain.User{
			NewPassword: &newPassword,
		}
	)

	j, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/password/forgot/%s", jwt), strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
		"email":    users[0].Email,
		"password": users[0].Password,
	}, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, usr)
}

func TestForgotPasswordConfirmNewPasswordInvalid(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	users, err := dbfixture.SeedActiveUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	var (
		userRepo    = repository.NewUserSqlxRepository(dbConn)
		newPassword = ""
		jwt         = createJWT(users[0], time.Minute*2)
		resp        domain.Response
		reqBody     = domain.User{
			NewPassword: &newPassword,
		}
	)

	t.Run("failed", func(t *testing.T) {
		j, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/password/forgot/%s", jwt), strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("x-access-token", jwt)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Message)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email":    users[0].Email,
			"password": users[0].Password,
		}, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
	})
}

func TestForgotPasswordConfirmSuccess(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	users, err := dbfixture.SeedUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	var (
		userRepo    = repository.NewUserSqlxRepository(dbConn)
		newPassword = "newpassword"
		jwt         = createJWT(users[0], time.Minute*2)
		resp        domain.Response
		reqBody     = domain.User{
			NewPassword: &newPassword,
		}
	)

	t.Run("make user active", func(t *testing.T) {
		err := makeUserActive(&users[0])
		assert.NoError(t, err)
		jwt = createJWT(users[0], time.Minute*2)
	})

	t.Run("success", func(t *testing.T) {
		j, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/password/forgot/%s", jwt), strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("x-access-token", jwt)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Message)
		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email":    users[0].Email,
			"password": reqBody.NewPassword,
		}, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
	})

}
