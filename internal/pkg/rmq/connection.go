package rmq

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Connection collects all of necessary field for connection to rabbitmq
type Connection struct {
	URL string
}

// NewConnection will create new an Connection object
func NewConnection(url string) *Connection {
	c := new(Connection)
	c.URL = url
	return c
}

// Dial to obtain rabbitmq connection
func (c *Connection) Dial() *amqp.Connection {
	for {
		log.Printf("Connecting to rabbitmq on %s\n", c.URL)
		conn, err := amqp.Dial(c.URL)

		if err == nil {
			return conn
		}

		logError("Connection to rabbitmq failed. Retrying in 1 sec... ", err)
		time.Sleep(1000 * time.Millisecond)
	}
}

// NotifyClose /
func (c *Connection) NotifyClose(conn *amqp.Connection, errCh chan *amqp.Error) {
	conn.NotifyClose(errCh)
}

// ErrorChannel /
func (c *Connection) ErrorChannel(conn *amqp.Connection) chan *amqp.Error {
	errCh := make(chan *amqp.Error)
	return errCh
}
