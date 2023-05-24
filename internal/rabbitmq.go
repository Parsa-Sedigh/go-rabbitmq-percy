package internal

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitClient is a wrapper around official amqp client which we use to add a bit more functionality
type RabbitClient struct {
	// The connection used by the client
	conn *amqp.Connection

	// Channel is used to process/send messages
	ch *amqp.Channel
}

func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, host, vhost))
}

func NewRabbitMQClient(conn *amqp.Connection) (RabbitClient, error) {
	/* take the connection and spawn a channel from it and this channel will be used for this new rabbit client. This allows us to reuse
	the connection with multiple rabbitmq clients.*/

	ch, err := conn.Channel()
	if err != nil {
		return RabbitClient{}, err
	}

	ch.QueueDeclare()

	return RabbitClient{
		conn: conn,
		ch:   ch,
	}, nil
}

// Close closes the channel for a particular client.
func (rc RabbitClient) Close() error {
	// Note: We only close the channel and we don't want to close the connection yet because we would have other clients using the same connection.
	return rc.ch.Close()
}
