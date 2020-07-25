package rmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// MsgCons type is for registered consumer
type MsgCons func(amqp.Delivery)

// Queue will provide few method to interact with rabbitmq
type Queue interface {
	Consume(consumer MsgCons)
	Publish(message string, routingKey string, header map[string]interface{}) error
	GetQueueName() string
}

// queue collection
type queue struct {
	name           string
	exchange       Exchange
	routingKey     []string
	connection     *amqp.Connection
	channel        *amqp.Channel
	consumerThread uint16
}

// NewQueue will create new an queue object
func NewQueue(
	qName string,
	conn *amqp.Connection,
	exchange Exchange,
	routingKey []string,
	isForConsume bool,
	isForPublish bool,
) Queue {
	q := new(queue)

	q.name = qName
	q.connection = conn
	q.exchange = exchange
	q.routingKey = routingKey
	q.channel = q.openChannel()

	if isForConsume {
		q.declareExchange()
		q.declareQueue()
		q.bindQueue()
	}

	return q
}

func (q *queue) openChannel() *amqp.Channel {
	channel, err := q.connection.Channel()
	logError("Opening channel failed", err)
	return channel
}

func (q *queue) declareExchange() {
	err := q.channel.ExchangeDeclare(
		q.exchange.ExcName,                    // name
		fmt.Sprintf("%s", q.exchange.ExcType), // type
		true,                                  // durable
		false,                                 // auto-deleted
		false,                                 // internal
		false,                                 // no-wait
		nil,                                   // arguments
	)
	logError("Exchange declaration failed ", err)
}

func (q *queue) declareQueue() {
	queueDlxName := q.name + ".dlx"

	arguments := make(amqp.Table)
	arguments["x-dead-letter-exchange"] = ""
	arguments["x-dead-letter-routing-key"] = queueDlxName

	_, err := q.channel.QueueDeclare(
		q.name,    // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		arguments, // arguments
	)
	logError("Queue declaration failed", err)

	_, err = q.channel.QueueDeclare(
		queueDlxName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	logError("Queue dlx declaration failed", err)
}

func (q *queue) bindQueue() {
	for _, r := range q.routingKey {
		err := q.channel.QueueBind(
			q.name,             // queue name
			r,                  // routing key
			q.exchange.ExcName, // exchange
			false,
			nil)
		logError(fmt.Sprintf("Bind a queue with routing key %s failed", r), err)
	}
}

func (q *queue) registerQueueConsumer() (<-chan amqp.Delivery, error) {
	msgs, err := q.channel.Consume(
		q.name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	logError("Consuming messages from queue failed", err)
	return msgs, err
}

func (q *queue) executeRabbitMessageConsumer(err error, consumer MsgCons, deliveries <-chan amqp.Delivery) {
	if err == nil {
		go func() {
			for delivery := range deliveries {
				log.Printf("New message from queue: %s, routing-key: %s", q.name, delivery.RoutingKey)
				consumer(delivery)
			}
		}()
	}
}

// Consume method for consume message from rabbitmq
func (q *queue) Consume(consumer MsgCons) {
	q.consumerThread++
	log.Println(fmt.Sprintf("Queue: %s. Registering consumer %d ...", q.name, q.consumerThread))
	deliveries, err := q.registerQueueConsumer()
	log.Println(fmt.Sprintf("Queue: %s. Consumer %d registered! Processing messages...", q.name, q.consumerThread))
	q.executeRabbitMessageConsumer(err, consumer, deliveries)
}

// Publish method for publishing message to rabbitmq
func (q *queue) Publish(message string, routingKey string, headers map[string]interface{}) error {
	log.Println("Sending message...")
	if routingKey == "" {
		routingKey = q.routingKey[0]
	}
	return q.channel.Publish(
		q.exchange.ExcName, // exchange
		routingKey,         // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			Headers:     headers,
		})
}

// GetQueueName to get queue name
func (q *queue) GetQueueName() string {
	return q.name
}

// QueuePurge to close channel
func (q *queue) QueuePurge() (int, error) {
	return q.channel.QueuePurge(q.name, true)
}
