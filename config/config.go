package config

import (
	"github.com/sirupsen/logrus"
)

func logError(message string, err error) {
	if err != nil {
		logrus.Printf("%s: %s", message, err)
	}
}
