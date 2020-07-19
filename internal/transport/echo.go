package transport

import (
	"time"

	"github.com/wicaker/user/internal/middleware"
	"github.com/wicaker/user/internal/pkg/rmq"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/internal/usecase"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
)

// Echo server
func Echo(db *sqlx.DB, rmqQ []rmq.Queue) *echo.Echo {
	e := echo.New()
	middL := middleware.InitEchoMiddleware()
	e.Use(middL.MiddlewareLogging)
	e.Use(middL.CORS)

	timeoutContext := time.Duration(2) * time.Second

	userRepo := repository.NewUserSqlxRepository(db)
	userUcase := usecase.NewUserUsecase(timeoutContext, userRepo)
	NewUserHandler(e, rmqQ, userUcase)

	return e
}
