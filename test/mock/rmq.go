package mock

import (
	"github.com/wicaker/user/internal/pkg/rmq"
)

// Message to accommodate published message include other stuff
type Message struct {
	Message    string
	RoutingKey string
	Headers    map[string]interface{}
}

type queue struct {
	name    string
	Message *Message
}

// NewMockQueueRMQ /
func NewMockQueueRMQ(name string, message *Message) rmq.Queue {
	return &queue{
		name:    name,
		Message: message,
	}
}

func (q *queue) Consume(consumer rmq.MsgCons) {
}

// Publish method for publishing message to rabbitmq
func (q *queue) Publish(message string, routingKey string, headers map[string]interface{}) error {
	*q.Message = Message{
		Message:    message,
		RoutingKey: routingKey,
		Headers:    headers,
	}
	return nil
}

// GetQueueName /
func (q *queue) GetQueueName() string {
	return q.name
}
