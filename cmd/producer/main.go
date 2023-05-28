package main

import (
	"context"
	"github.com/Parsa-Sedigh/go-rabbitmq-percy/internal"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func main() {
	conn, err := internal.ConnectRabbitMQ("parsa", "secret", "localhost:5672", "customers")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client, err := internal.NewRabbitMQClient(conn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := client.Send(ctx, "customer_events", "customers.created.us", amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(`A cool message between services`),
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
}
