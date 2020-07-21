package integration_test

import (
	"context"
	"encoding/json"
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

func TestChangeEmailUserInactive(t *testing.T) {
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
		userRepo = repository.NewUserSqlxRepository(dbConn)
		userOld  = users[0]
		userNew  = domain.User{
			Email:    "new@mail.com",
			Password: "Password1",
		}
		jwt  = createJWT(userOld, time.Minute*2)
		resp domain.Response
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	assert.Empty(t, resp.Data)

	u, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
		"email": &userNew.Email,
	}, nil)
	assert.NoError(t, err)
	assert.Empty(t, u)

}
func TestChangeEmailReqNotProvideToken(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	var (
		userNew = domain.User{
			Email:    "new@mail.com",
			Password: "Password1",
		}
		resp domain.Response
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestChangeEmailTokenExpired(t *testing.T) {
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
		userOld = users[0]
		userNew = domain.User{
			Email:    "new@mail.com",
			Password: "Password1",
		}
		jwt  = createJWT(userOld, time.Millisecond*1)
		resp domain.Response
	)

	time.Sleep(time.Second * 1) //wait until token expired

	j, _ := json.Marshal(userNew)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestChangeEmailUserNotFound(t *testing.T) {
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
		mockUserOld = domain.User{
			UUID:     users[0].UUID,
			Email:    "random@mail.com",
			Password: users[0].Password,
		}
		userNew = domain.User{
			Email:    "new@mail.com",
			Password: "Password1",
		}
		jwt  = createJWT(mockUserOld, time.Minute*1)
		resp domain.Response
	)

	j, _ := json.Marshal(userNew)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestChangeEmailPasswordNotMatch(t *testing.T) {
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
		userOld = users[0]
		userNew = domain.User{
			Email:    "new@mail.com",
			Password: "random",
		}
		jwt  = createJWT(userOld, time.Minute*5)
		resp domain.Response
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
}

func TestChangeEmailSuccess(t *testing.T) {
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
		userRepo = repository.NewUserSqlxRepository(dbConn)
		userOld  = users[0]
		userNew  = domain.User{
			Email:    "new@mail.com",
			Password: "Password1",
		}
		jwt  = createJWT(userOld, time.Minute*2)
		resp domain.Response
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/email", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
	assert.Empty(t, resp.Data)

	u, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
		"email": &userNew.Email,
	}, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}
