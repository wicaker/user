package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/test/dbfixture"
)

func TestForgotPasswordEmailNotAppropriate(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	var (
		user = &domain.User{
			Email: "test@gmail.com",
		}
		resp domain.Response
	)

	j, _ := json.Marshal(user)

	req, err := http.NewRequest(http.MethodPut, "/user/password/forgot", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestForgotPasswordEmailNotActive(t *testing.T) {
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
			Email: users[0].Email,
		}
		resp domain.Response
	)

	j, _ := json.Marshal(user)

	req, err := http.NewRequest(http.MethodPut, "/user/password/forgot", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestForgotPasswordSuccess(t *testing.T) {
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
		reqBody = domain.User{
			Email: users[0].Email,
		}
		resp domain.Response
	)

	j, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPut, "/user/password/forgot", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)

	msg := getMessageInMq()
	parsedToken, err := middleware.JwtVerify(msg.Token)
	assert.NoError(t, err)
	assert.Equal(t, users[0].Email, msg.EmailDestination)
	assert.NotEmpty(t, parsedToken.Salt)
}
