package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/wicaker/user/config"
	"github.com/wicaker/user/internal/transport"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"at": time.Now().Format("2006-01-02 15:04:05"),
		}).Println(err)
	}
}

func main() {
	var (
		sqlxConf   = config.NewSqlx()
		dbConn     *sqlx.DB
		rabbitConn = config.NewRabbitmq()
		errChan    = make(chan error)
	)
	defer close(errChan)
	defer func() {
		err := rabbitConn.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"at": time.Now().Format("2006-01-02 15:04:05"),
			}).Errorln(err)
		}
	}()
	defer func() {
		err := sqlxConf.Close(dbConn)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"at": time.Now().Format("2006-01-02 15:04:05"),
			}).Errorln(err)
		}
	}()

	// open db connection
	dbConn, err := sqlxConf.Open()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"at": time.Now().Format("2006-01-02 15:04:05"),
		}).Fatal(err)
	}

	go func() {
		for {
			errServerClosed := make(chan error)
			eServer := transport.Echo(dbConn, rabbitConn.Queue)
			srv := &http.Server{
				Addr:         ":" + os.Getenv("SERVER_ECHO_PORT"),
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
				Handler:      eServer,
			}
			logrus.WithFields(logrus.Fields{
				"at": time.Now().Format("2006-01-02 15:04:05"),
			}).Printf("Starting echo server on port :%s...", os.Getenv("SERVER_ECHO_PORT"))

			// start http server
			go func() {
				err := srv.ListenAndServe()
				if err != nil {
					if err == http.ErrServerClosed {
						errServerClosed <- err
					} else {
						errChan <- err
					}
				}
			}()

			// reconnect when rabbitmq server terminated accidentally
			err = <-rabbitConn.ErrorChannel
			if !rabbitConn.IsClose {
				rabbitConn.Reconnect(err)
			} else {
				errChan <- err
			}

			// shutdown http server gracefully
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				errChan <- err
			}

			logrus.WithFields(logrus.Fields{
				"at": time.Now().Format("2006-01-02 15:04:05"),
			}).Println(<-errServerClosed)
		}
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	logrus.WithFields(logrus.Fields{
		"at": time.Now().Format("2006-01-02 15:04:05"),
	}).Errorln("terminated ", <-errChan)
}
