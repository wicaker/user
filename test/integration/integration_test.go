package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang-migrate/migrate"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/wicaker/user/config"
	"github.com/wicaker/user/internal/domain"
	"github.com/wicaker/user/internal/pkg/rmq"
	"github.com/wicaker/user/internal/repository"
	"github.com/wicaker/user/internal/transport"
	"github.com/wicaker/user/test/dbfixture"
	"github.com/wicaker/user/test/mock"
)

type messageInMq struct {
	EmailDestination string `json:"email_destination"`
	Token            string `json:"token"`
}

var (
	dbConn           *sqlx.DB
	migrate_down     string
	api              *echo.Echo
	publishedMessage mock.Message
	listrmq          []rmq.Queue
)

func init() {
	os.Setenv("DATABASE_URL", "postgres://root:root@localhost:5432/microservice_user?sslmode=disable")
	os.Setenv("RABBITMQ_SERVER", "amqp://guest:guest@localhost:5672/")
	migrate_down = os.Getenv("migrate_down")
}

func TestMain(m *testing.M) {
	var (
		err      error
		sqlxConf = config.NewSqlx()
	)

	// open db connection
	dbConn, err = sqlxConf.Open()
	if err != nil {
		log.Fatal(err)
	}

	// registering queue or channel
	registerMockQueue()

	// registering transport
	api = transport.Echo(dbConn, listrmq)

	// migrate up
	err = sqlxConf.MigrateUp("file://../../migrations")
	if err != nil {
		if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", migrate.ErrNoChange) {
			log.Fatal(err)
		} else {
			if err := dbfixture.Truncate(dbConn); err != nil {
				log.Fatal(err)
			}
		}
	}

	// runs the tests
	code := m.Run()

	// migrate down
	if migrate_down == "true" {
		err = sqlxConf.MigrateDown()
		if err != nil {
			log.Fatal(err)
		}
	}

	// close db connection
	err = sqlxConf.Close(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func getMessageInMq() messageInMq {
	defer func() {
		publishedMessage = mock.Message{}
	}()

	message := messageInMq{}

	if publishedMessage.Message != "" {
		err := json.Unmarshal([]byte(publishedMessage.Message), &message)
		if err != nil {
			log.Fatal(err)
		}
	}

	return message
}

func createJWT(user domain.User, exp time.Duration) string {
	expiresAt := time.Now().Add(exp).Unix()
	tk := &domain.JWToken{
		UUID:  user.UUID,
		Email: user.Email,
		Salt:  user.Salt,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Fatal(err)
	}

	return tokenString
}

func makeUserActive(user *domain.User) error {
	userRepo := repository.NewUserSqlxRepository(dbConn)

	user.IsActive = true

	_, err := userRepo.Update(context.TODO(), user)
	return err
}

func registerMockQueue() {
	listrmq = append(listrmq, mock.NewMockQueueRMQ("publish-user-register", &publishedMessage))
	listrmq = append(listrmq, mock.NewMockQueueRMQ("publish-user-change-password", &publishedMessage))
	listrmq = append(listrmq, mock.NewMockQueueRMQ("publish-user-forgot-password", &publishedMessage))
}
