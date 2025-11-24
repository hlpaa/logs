package main

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func QueryConsumer(conn *amqp091.Connection, userID string) {
	// Cria um canal
	queryChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer queryChannel.Close()

	queryQueue, err := queryChannel.QueueDeclare(
		"query." + userID,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Consome mensagens
	messages, err := queryChannel.Consume(
		queryQueue.Name, // fila
		"",     // consumer
		true,   // auto-ack (confirmação automática)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatal(err)
	}

	// Lê mensagens em um loop
	forever := make(chan bool)

	go func() {
		for received := range messages {
			var query Log
			err := json.Unmarshal(received.Body, &query)
			if err != nil {
				log.Print(err)
				return
			}

			log.Print(query)
		}
	}()

	log.Printf("Aguardando queries...")
	<-forever
}
