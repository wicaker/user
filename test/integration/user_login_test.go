package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/test/dbfixture"
)

func TestLoginUserStillInactive(t *testing.T) {
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
		user = domain.User{
			Email:    users[0].Email,
			Password: "Password1",
		}
		resp domain.Response
	)

	j, err := json.Marshal(user)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	assert.NotEmpty(t, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestLoginUserSuccess(t *testing.T) {
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
		user = domain.User{
			Email:    users[0].Email,
			Password: "Password1",
		}
		resp domain.Response
	)

	t.Run("make user active", func(t *testing.T) {
		err := makeUserActive(&users[0])
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		j, err := json.Marshal(user)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NotEmpty(t, resp.Data["token"])
		assert.Nil(t, resp.Errors)
	})
}

func TestLoginUserFailed(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	users, err := dbfixture.SeedUsers(dbConn, 1)
	if err != nil {
		t.Error(err)
	}

	t.Run("make user active", func(t *testing.T) {
		err := makeUserActive(&users[0])
		assert.NoError(t, err)
	})

	t.Run("failed, email not found", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{
				Email:    "test@mail.com",
				Password: "123",
			}
		)

		j, _ := json.Marshal(mockUser)

		req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err := json.Unmarshal(w.Body.Bytes(), &resp)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Equal(t, domain.ErrUserNotFound.Error(), resp.Message)
	})

	t.Run("failed, wrong password", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{
				Email:    users[0].Email,
				Password: "1",
			}
		)

		j, _ := json.Marshal(mockUser)

		req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err := json.Unmarshal(w.Body.Bytes(), &resp)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
		assert.Equal(t, domain.ErrWrongPassword.Error(), resp.Message)
	})

	t.Run("failed because not contain email and password", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{}
		)

		j, _ := json.Marshal(mockUser)

		req, err := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Validation error", resp.Message)
		assert.Equal(t, 2, len(resp.Errors))
	})
}
