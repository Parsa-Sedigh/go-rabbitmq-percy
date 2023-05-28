package internal

import (
	"context"
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

	// enable the confirm mode
	if err := ch.Confirm(false); err != nil {
		return RabbitClient{}, err
	}

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

// CreateQueue is a wrapper around QueueDeclare. So we don't allow the customization of all the params of QueueDeclare. CreateQueue
// will create a new queue based on given cfgs
func (rc RabbitClient) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
	q, err := rc.ch.QueueDeclare(queueName, durable, autodelete, false, false, nil)

	if err != nil {
		return amqp.Queue{}, err
	}

	return q, err
}

// CreateBinding will bind the current channel to the given exchange using the routingKey provided
// `name` param is the name of the queue that we wanna bind
func (rc RabbitClient) CreateBinding(name, binding, exchange string) error {
	return rc.ch.QueueBind(name, binding, exchange, false, nil)
}

// Send is a wrapper function to publish payloads onto exchange an exchange with the given routingKey. options is the actual message that we wanna send
func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	//return rc.ch.PublishWithContext(
	//	ctx,
	//	exchange,
	//	routingKey,
	//	// mandatory is used to determine if an error should be returned upon failure
	//	true,
	//	// immediate
	//	false,
	//	options,
	//)

	///////////////////

	confirmation, err := rc.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		routingKey,
		// mandatory is used to determine if an error should be returned upon failure
		true,
		// immediate
		false,
		options,
	)
	if err != nil {
		return err
	}

	// this will block until we receive information about the sent message
	confirmation.Wait()

	return nil
}

// Consume is used to consume a queue
func (rc RabbitClient) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}
