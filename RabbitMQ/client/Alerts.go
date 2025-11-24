package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type ErrorPeak struct {
	UserID string `json:"user_id"`
	Count int `json:"count"`
	Interval time.Duration `json:"interval"`
}

func AlertsConsumer(conn *amqp091.Connection, userID string) {
	// Cria um canal
	alertChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer alertChannel.Close()

	alertQueue, err := alertChannel.QueueDeclare(
		"error." + userID,
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
	messages, err := alertChannel.Consume(
		alertQueue.Name, // fila
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
			var alert ErrorPeak
			err := json.Unmarshal(received.Body, &alert)
			if err != nil {
				log.Print(err)
				return
			}

			log.Print(alert)
		}
	}()

	log.Printf("Aguardando alertas...")
	<-forever
}
