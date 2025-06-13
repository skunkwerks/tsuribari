package queue

import (
	"encoding/json"

	"github.com/streadway/amqp"

	"tsuribari/internal/models"
)

type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func NewRabbitMQ(url, exchange, queue string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue,
		queue,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (r *RabbitMQ) PublishWorkflow(workflow *models.Workflow) error {
	body, err := json.Marshal(workflow)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		r.exchange,
		r.queue,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
