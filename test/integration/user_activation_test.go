package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/test/dbfixture"
)

func TestActivationHandler(t *testing.T) {
	defer func() {
		if err := dbfixture.Truncate(dbConn); err != nil {
			t.Errorf("error truncating test database tables: %v", err)
		}
	}()

	var (
		user = domain.User{
			Email:    "testactivation@mail.com",
			Password: "testactivation",
		}
		msg messageInMq
	)

	// register a new user
	t.Run("register a new user", func(t *testing.T) {
		j, _ := json.Marshal(user)
		req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(string(j)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)

		assert.Equal(t, "user.register", publishedMessage.RoutingKey)
		msg = getMessageInMq()
		assert.Equal(t, "testactivation@mail.com", msg.EmailDestination)
		assert.NotEmpty(t, msg.Token)
	})

	t.Run("failed, because not contain key params", func(t *testing.T) {
		var (
			resp domain.Response
		)

		j, _ := json.Marshal(user)

		req, err := http.NewRequest(http.MethodPut, "/user/activation", strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("success,  because user inactive", func(t *testing.T) {
		var (
			resp domain.Response
		)

		j, err := json.Marshal(user)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/activation/%s", msg.Token), strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
		assert.Empty(t, resp.Data)
		assert.Nil(t, resp.Errors)
	})

	t.Run("failed, because user already active", func(t *testing.T) {
		var (
			resp domain.Response
		)

		j, _ := json.Marshal(user)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/activation/%s", msg.Token), strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})
}
