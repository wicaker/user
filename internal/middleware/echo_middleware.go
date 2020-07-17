package middleware

import (
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// EchoMiddleware represent the data-struct for middleware
type EchoMiddleware struct {
	// another stuff , may be needed by middleware
}

// InitEchoMiddleware intialize the middleware
func InitEchoMiddleware() *EchoMiddleware {
	return &EchoMiddleware{}
}

// CORS will handle the CORS middleware
func (m *EchoMiddleware) CORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		return next(c)
	}
}

// MiddlewareLogging for logging
func (m *EchoMiddleware) MiddlewareLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		makeLogEntry(c).Info("incoming request")
		return next(c)
	}
}

func makeLogEntry(c echo.Context) *logrus.Entry {
	if c == nil {
		return logrus.WithFields(logrus.Fields{
			"at": time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return logrus.WithFields(logrus.Fields{
		"at":     time.Now().Format("2006-01-02 15:04:05"),
		"method": c.Request().Method,
		"uri":    c.Request().URL.String(),
		"ip":     c.Request().RemoteAddr,
	})
}
