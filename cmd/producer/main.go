package main

import (
	"github.com/Parsa-Sedigh/go-rabbitmq-percy/internal"
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

	time.Sleep(10 * time.Second)

	log.Println("client: ", client)
}
