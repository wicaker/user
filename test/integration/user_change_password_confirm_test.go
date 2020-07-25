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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/internal/usecase"
	"github.com/wicaker/user/test/dbfixture"
)

func TestUserChangePasswordConfirmSuccess(t *testing.T) {
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
		userOld      = users[0]
		userNew      = domain.User{
			Email:       users[0].Email,
			Password:    "Password1",
			NewPassword: &newPassoword,
		}
		msg messageInMq
	)

	t.Run("success change password, but not yet to confirm", func(t *testing.T) {
		var (
			jwt         = createJWT(userOld, time.Minute*2)
			resp        domain.Response
			mockUserOld = domain.User{
				Email:    userOld.Email,
				Password: "Password1",
			}
			mockUserNew = domain.User{
				Email:    userNew.Email,
				Password: *userNew.NewPassword,
			}
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
		jwt, err = userUsecase.Login(context.TODO(), &mockUserOld)
		assert.NoError(t, err)
		assert.NotEmpty(t, jwt)

		jwt, err = userUsecase.Login(context.TODO(), &mockUserNew)
		assert.Error(t, err)
		assert.Empty(t, jwt)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": userOld.Email,
		}, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
		assert.Equal(t, usr.Email, userNew.Email)
		assert.NotEmpty(t, usr.NewPassword)

		msg = getMessageInMq()
		parsedToken, err := middleware.JwtVerify(msg.Token)
		assert.NoError(t, err)
		assert.Equal(t, users[0].Email, msg.EmailDestination)
		assert.Equal(t, usr.Salt, parsedToken.Salt)
	})

	t.Run("success confirm change password", func(t *testing.T) {
		var (
			mockUser    domain.User
			resp        domain.Response
			mockUserNew = domain.User{
				Email:    userNew.Email,
				Password: *userNew.NewPassword,
			}
		)

		j, err := json.Marshal(mockUser)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/user/password/change/%s", msg.Token), strings.NewReader(string(j)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
		assert.Empty(t, resp.Data)

		jwt, err := userUsecase.Login(context.TODO(), &mockUserNew)
		assert.NoError(t, err)
		assert.NotEmpty(t, jwt)

		usr, err := userRepo.FindOneBy(context.TODO(), map[string]interface{}{
			"email": userOld.Email,
		}, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
		assert.Equal(t, usr.Email, userNew.Email)
		assert.Equal(t, *usr.NewPassword, usr.Password)
	})
}
