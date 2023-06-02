package main

import (
	"context"
	"fmt"
	"github.com/Parsa-Sedigh/go-rabbitmq-percy/internal"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"time"
)

func main() {
	// TODO: Use env vars for empty strings
	conn, err := internal.ConnectRabbitMQ("parsa", "secret", "localhost:5671", "customers",
		os.Getenv("caCert"), os.Getenv("clientCert"), os.Getenv("clientKey"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// all consuming will be done on this connection
	consumeConn, err := internal.ConnectRabbitMQ("parsa", "secret", "localhost:5671", "customers",
		os.Getenv("caCert"), os.Getenv("clientCert"), os.Getenv("clientKey"))
	if err != nil {
		panic(err)
	}
	defer consumeConn.Close()

	client, err := internal.NewRabbitMQClient(conn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	consumeClient, err := internal.NewRabbitMQClient(consumeConn)
	if err != nil {
		panic(err)
	}
	defer consumeClient.Close()

	////// Note: In pub/sub, producer doesn't create the queue.
	//if err := client.CreateQueue("customers_created", true, false); err != nil {
	//	panic(err)
	//}
	//if err := client.CreateQueue("customers_test", false, true); err != nil {
	//	panic(err)
	//}

	//if err := client.CreateBinding("customers_created", "customers.created.*", "customer_events"); err != nil {
	//	panic(err)
	//}
	//if err := client.CreateBinding("customers_test", "customers.*", "customer_events"); err != nil {
	//	panic(err)
	//}

	/////// For RPC section ///////
	// we use an unnamed queue so rabbitmq will create a random name
	queue, err := consumeClient.CreateQueue("", true, true)
	if err != nil {
		panic(err)
	}

	// bind the queue created by consumeClient
	/* The binding would have the name of the queue because this is a direct exchange */
	if err := consumeClient.CreateBinding(queue.Name, queue.Name, "customer_callbacks"); err != nil {
		panic(err)
	}

	// consuming
	/* Let's assume we have a customer-api push a new member that registered and then it expects a callback to notify the user if the
	registration was successful or not.*/
	messageBus, err := consumeClient.Consume(queue.Name, "customer-api", true)
	if err != nil {
		panic(err)
	}

	go func() {
		/* For demonstration reasons, we're not gonna create multiple goroutines here. We're just gonna create a single-threaded reader from the
		messageBus. To create one goroutine for each incoming message, see the consuming code for the second queue later in this file.*/
		for message := range messageBus {
			log.Printf("Message Callback %s\n", message.CorrelationId)
		}
	}()

	/////// ///////

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := client.Send(ctx, "customer_events", "customers.created.us", amqp.Publishing{
			ContentType:   "text/plain",
			DeliveryMode:  amqp.Persistent,
			ReplyTo:       queue.Name,
			CorrelationId: fmt.Sprintf("customer_created_%d", i),
			Body:          []byte(`A cool message between services`),
		}); err != nil {
			panic(err)
		}

		// sending a transient message
		//if err := client.Send(ctx, "customer_events", "customers.test", amqp.Publishing{
		//	ContentType:  "text/plain",
		//	DeliveryMode: amqp.Transient,
		//	Body:         []byte(`A un-durable message between services`),
		//}); err != nil {
		//	panic(err)
		//}
	}

	//time.Sleep(10 * time.Second)

	log.Println("client: ", client)

	var blocking chan struct{}

	<-blocking
}
