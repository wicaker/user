package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/test/dbfixture"
)

func TestRegisterHandlerSuccess(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	var (
		userRepo = repository.NewUserSqlxRepository(dbConn)
		mockUser = &domain.User{
			Email:    "register1@mail.com",
			Password: "123",
		}
	)

	t.Run("success", func(t *testing.T) {
		var (
			resp domain.Response
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
		assert.NotEmpty(t, resp.Message)
		assert.Nil(t, resp.Data)
		assert.Nil(t, resp.Errors)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": &mockUser.Email,
		}, nil)
		assert.False(t, usr.IsActive)
		assert.NotEmpty(t, usr.Salt)

		msg := getMessageInMq()
		parsedToken, err := middleware.JwtVerify(msg.Token)
		assert.NoError(t, err)
		assert.Equal(t, "register1@mail.com", msg.EmailDestination)
		assert.Equal(t, usr.Salt, parsedToken.Salt)
		assert.NotEmpty(t, parsedToken.UUID)
	})

	t.Run("success, because user still inactive", func(t *testing.T) {
		var (
			resp domain.Response
		)

		mockUser.Password = "12345"
		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
		assert.NotEmpty(t, resp.Message)
		assert.Nil(t, resp.Data)
		assert.Nil(t, resp.Errors)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": &mockUser.Email,
		}, nil)
		assert.False(t, usr.IsActive)
		assert.NotEmpty(t, usr.Salt)
		mockUser.UUID = usr.UUID

		msg := getMessageInMq()
		parsedToken, err := middleware.JwtVerify(msg.Token)
		assert.NoError(t, err)
		assert.Equal(t, "register1@mail.com", msg.EmailDestination)
		assert.Equal(t, usr.Salt, parsedToken.Salt)
		assert.NotEmpty(t, parsedToken.UUID)
	})
}

func TestRegisterHandlerFailed(t *testing.T) {
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

	t.Run("failed because user already exist", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = &domain.User{
				Email:    users[0].Email,
				Password: "Password123",
			}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
		assert.Equal(t, domain.ErrUserAlreadyExist.Error(), resp.Message)
		assert.Nil(t, resp.Errors)
		assert.Nil(t, resp.Data)

		msg := getMessageInMq()
		assert.Empty(t, msg)
	})

	t.Run("failed because not contain email", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{
				Password: "123",
			}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Validation error", resp.Message)
		assert.Equal(t, 1, len(resp.Errors))
		assert.Nil(t, resp.Data)

		msg := getMessageInMq()
		assert.Empty(t, msg)
	})

	t.Run("failed because not contain password", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{
				Email: "123@test.com",
			}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Validation error", resp.Message)
		assert.Equal(t, 1, len(resp.Errors))
		assert.Nil(t, resp.Data)

		msg := getMessageInMq()
		assert.Empty(t, msg)
	})

	t.Run("failed because wrong format email", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{
				Email:    "123test.com",
				Password: "123",
			}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Validation error", resp.Message)
		assert.Equal(t, 1, len(resp.Errors))
		assert.Nil(t, resp.Data)

		msg := getMessageInMq()
		assert.Empty(t, msg)
	})

	t.Run("failed because not contain email and password", func(t *testing.T) {
		var (
			resp     domain.Response
			mockUser = domain.User{}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Validation error", resp.Message)
		assert.Equal(t, 2, len(resp.Errors))
		assert.Nil(t, resp.Data)

		msg := getMessageInMq()
		assert.Empty(t, msg)
	})
}
