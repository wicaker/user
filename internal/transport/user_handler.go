package transport

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/internal/pkg/rmq"
)

// UserHandler represent the httphandler for user
type UserHandler struct {
	UserUsecase                domain.UserUsecase
	queuePublishRegister       rmq.Queue
	queuePublishChangePassword rmq.Queue
	queuePublishForgotPassword rmq.Queue
}

// NewUserHandler will initialize the user endpoint
func NewUserHandler(e *echo.Echo, rmqQueue []rmq.Queue, u domain.UserUsecase) {
	handler := &UserHandler{
		UserUsecase: u,
	}

	for _, rmqQ := range rmqQueue {
		switch name := rmqQ.GetQueueName(); name {
		case "publish-user-register":
			handler.queuePublishRegister = rmqQ
		case "publish-user-change-password":
			handler.queuePublishChangePassword = rmqQ
		case "publish-user-forgot-password":
			handler.queuePublishForgotPassword = rmqQ
		}
	}

	e.POST("/user/register", handler.Register)
	e.POST("/user/login", handler.Login)
	e.PUT("/user/activation/:token", handler.Activation)
	e.PUT("/user/email", handler.ChangeEmail)
	e.PUT("/user/password/change", handler.ChangePassword)
	e.PUT("/user/password/change/:token", handler.PasswordConfirm)
	e.PUT("/user/password/forgot", handler.ForgotPasswordRequest)
	e.PUT("/user/password/forgot/:token", handler.ForgotPasswordConfirm)
}

// Register will handle register request
func (uh *UserHandler) Register(c echo.Context) error {
	var user domain.User

	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.Response{Message: err.Error()})
	}

	if ok, err := middleware.Validate(&user); !ok {
		return c.JSON(http.StatusBadRequest, domain.Response{Message: "Validation error", Errors: err})
	}

	ctx := c.Request().Context()
	if ctx == nil {
		ctx = context.Background()
	}

	token, err := uh.UserUsecase.Register(ctx, &user)
	if err != nil {
		return c.JSON(domain.GetStatusCode(err), domain.Response{Message: err.Error()})
	}

	rabbitMessage := fmt.Sprintf(`{"email_destination":"%s","token":"%s"}`, user.Email, token)
	err = uh.queuePublishRegister.Publish(rabbitMessage, "user.register", make(map[string]interface{}))
	if err != nil {
		log.Println(err)
	}

	return c.JSON(http.StatusCreated, domain.Response{Message: "Successfully register new user. Please confirm your email address!"})
}

// Login will handle login request
func (uh *UserHandler) Login(c echo.Context) error {
	var user domain.User

	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, domain.Response{Message: err.Error()})
	}

	if ok, err := middleware.Validate(&user); !ok {
		return c.JSON(http.StatusBadRequest, domain.Response{Message: "Validation error", Errors: err})
	}

	ctx := c.Request().Context()
	if ctx == nil {
		ctx = context.Background()
	}

	token, err := uh.UserUsecase.Login(ctx, &user)
	if err != nil {
		return c.JSON(domain.GetStatusCode(err), domain.Response{Message: err.Error()})
	}

	respData := map[string]interface{}{
		"token": token,
	}

	return c.JSON(http.StatusOK, domain.Response{Message: "Login successfully", Data: respData})
}

// Activation will handle activation request for user first time register
func (uh *UserHandler) Activation(c echo.Context) error {
	ctx := c.Request().Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// get token
	token := c.Param("token")
	parsedToken, err := middleware.JwtVerify(token)
	if err != nil {
		return c.JSON(domain.GetStatusCode(err), domain.Response{Message: err.Error()})
	}

	err = uh.UserUsecase.Activation(ctx, *parsedToken)
	if err != nil {
		return c.JSON(domain.GetStatusCode(err), domain.Response{Message: err.Error()})
	}

	return c.JSON(http.StatusNoContent, domain.Response{Message: "Activation request successfully"})
}

// ChangeEmail will handle change email request
func (uh *UserHandler) ChangeEmail(c echo.Context) error {
	return nil
}

// ChangePassword will handle change password request
func (uh *UserHandler) ChangePassword(c echo.Context) error {
	return nil
}

// PasswordConfirm will handle confirmation of change password request
func (uh *UserHandler) PasswordConfirm(c echo.Context) error {
	return nil
}

// ForgotPasswordRequest will handle forgot password request
func (uh *UserHandler) ForgotPasswordRequest(c echo.Context) error {
	return nil
}

// ForgotPasswordConfirm will handle forgot password request
func (uh *UserHandler) ForgotPasswordConfirm(c echo.Context) error {
	return nil
}
