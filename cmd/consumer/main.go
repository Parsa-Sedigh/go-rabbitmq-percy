package main

import (
	"context"
	"github.com/Parsa-Sedigh/go-rabbitmq-percy/internal"
	"golang.org/x/sync/errgroup"
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

	// we named this consumer, email-service because this consumer will receive them and send emails
	//messageBus, err := client.Consume("customers_created", "email-service", false)
	//if err != nil {
	//	panic(err)
	//}

	queue, err := client.CreateQueue("", true, true)
	if err != nil {
		panic(err)
	}

	// the binding is empty string because it's fanout so we don't care about the binding
	/* bound the queue using a fanout exchange named "customer_events".
	This will allow this client to receive all the events that are produced.*/
	if err := client.CreateBinding(queue.Name, "", "customer_events"); err != nil {
		panic(err)
	}

	// start consuming the queue
	messageBus, err := client.Consume(queue.Name, "email-service", false)
	if err != nil {
		panic(err)
	}

	////////////// without using errgroups //////////////
	var blocking chan struct{}

	//go func() {
	//	for message := range messageBus {
	//		log.Printf("New Message: %v\n", message)
	//
	//		//if err := message.Ack(false); err != nil {
	//		//	log.Println("Acknowledge message failed: ", err)
	//		//	continue
	//		//}
	//
	//		/* The 2 next if blocks will make sure each message is first NACKed(yeah not practical) because Redelivered is set to false
	//		the first time and then next time we received that message, it's ACKed. So first time, we're nacking the message and the second time we
	//		received that message, we're acking it.*/
	//
	//		// this block is not a real-world thing that we wanna do!
	//		if !message.Redelivered {
	//			message.Nack(false, true)
	//			continue
	//		}
	//
	//		if err := message.Ack(false); err != nil {
	//			log.Println("Failed to ack message")
	//			continue
	//		}
	//
	//		log.Printf("Acknowledge message %s\n", message.MessageId)
	//	}
	//}()

	//	log.Println("Consuming, to close the program press CTRL+C")
	//
	//	// block forever
	//	<-blocking
	////////////// //////////////

	// set timeout for 15 seconds per task
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// errgroup allows us to set concurrent tasks
	// set a limit of 10 concurrent goroutines at the same time
	g.SetLimit(10)

	// this allows us to listen for multiple messages at the same time
	go func() {
		for message := range messageBus {
			/* In this approach instead of handling the message here, we will spawn a worker */

			// initialize a new variable to prevent overwriting the `message` variable
			msg := message
			g.Go(func() error {
				log.Printf("New Message: %v", msg)
				time.Sleep(10 * time.Second) // simulate a long work

				if err := msg.Ack(false); err != nil {
					log.Println("Ack message failed")
					return err
				}

				log.Printf("Acknowledged message %s\n", message.MessageId)

				return nil
			})
		}
	}()

	log.Println("Consuming, to close the program press CTRL+C")

	<-blocking
}
