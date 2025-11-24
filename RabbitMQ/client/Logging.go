package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type EventRequest struct {
	UserID    string `json:"user_id"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

func LoggingProducer(conn *amqp091.Connection, userID string) {
	// Cria um canal
	logChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer logChannel.Close()

	logQueue, err := logChannel.QueueDeclare(
		"event.log",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	options := []string{"INFO", "WARN", "ERROR"}
	n := len(options)

	for {
		time.Sleep(2 * time.Second)
		
		now := time.Now().Format(time.RFC3339)
		randomIndex := rand.IntN(n)
		severity := options[randomIndex]

		report := EventRequest{
			UserID: userID,
			Timestamp: now,
			Severity:  severity,
			Message:   fmt.Sprintf("[%s] Logging data from user %s", severity, userID),
		}
		data, err := json.Marshal(report)
		if err != nil {
			log.Printf("Error marshaling JSON: %s", err)
		}

		err = logChannel.Publish(
			"",     			// exchange
			logQueue.Name, 		// routing key
			false,  			// mandatory
			false,  			// immediate
			amqp091.Publishing{
				Body: data,
				ContentType: "application/json", 	// Important!
				DeliveryMode: amqp091.Persistent, 	// Make message persistent
			})
		if err != nil {
			log.Print(err.Error())
			continue
		}
	}
}
