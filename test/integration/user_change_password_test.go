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
	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/internal/usecase"
	"github.com/wicaker/user/test/dbfixture"
)

func TestUserChangePasswordUserNotActive(t *testing.T) {
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
		newPassoword = "newpassword"
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		resp domain.Response
		jwt  = createJWT(users[0], time.Minute*5)
	)

	j, _ := json.Marshal(userNew)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestUserChangePasswordReqNotProvideToken(t *testing.T) {
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
		newPassoword = "newpassword"
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		resp domain.Response
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestUserChangePasswordTokenExpired(t *testing.T) {
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
		newPassoword = "newpassword"
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		resp domain.Response
		jwt  = createJWT(users[0], time.Millisecond*1)
	)

	time.Sleep(time.Second * 1) //wait until token expired

	j, _ := json.Marshal(userNew)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestUserChangePasswordUserNotFound(t *testing.T) {
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
		newPassoword = "newpassword"
		userOld      = domain.User{
			UUID:     users[0].UUID,
			Email:    "random@mail.com",
			Password: "Password1",
		}
		userNew = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		jwt  = createJWT(userOld, time.Minute*5)
		resp domain.Response
	)

	j, _ := json.Marshal(userNew)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestUserChangePasswordOldPasswordNotMatch(t *testing.T) {
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
		newPassoword = "newpassword"
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "random",
			NewPassword: &newPassoword,
		}
		resp domain.Response
		jwt  = createJWT(users[0], time.Minute*1)
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-access-token", jwt)

	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	msg := getMessageInMq()
	assert.Empty(t, msg)
}

func TestUserChangePasswordSuccess(t *testing.T) {
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
		userRepo     = repository.NewUserSqlxRepository(dbConn)
		userUsecase  = usecase.NewUserUsecase(time.Duration(2)*time.Second, userRepo)
		newPassoword = "newpassword"
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		resp domain.Response
		jwt  = createJWT(users[0], time.Minute*1)
	)

	j, err := json.Marshal(userNew)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, "/user/password/change", strings.NewReader(string(j)))
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

	// because still using oldPassword, need email confirmation to activate new password
	jwt, err = userUsecase.Login(context.TODO(), &domain.User{Email: users[0].Email, Password: "Password1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, jwt)

	jwt, err = userUsecase.Login(context.TODO(), &domain.User{Email: userNew.Email, Password: *userNew.NewPassword})
	assert.Error(t, err)
	assert.Empty(t, jwt)

	usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
		"email": users[0].Email,
	}, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, usr)
	assert.Equal(t, usr.Email, userNew.Email)
	assert.NotEmpty(t, usr.NewPassword)

	msg := getMessageInMq()
	token, err := middleware.JwtVerify(msg.Token)
	assert.NoError(t, err)
	assert.Equal(t, users[0].Email, msg.EmailDestination)
	assert.Equal(t, usr.Salt, token.Salt)
}
