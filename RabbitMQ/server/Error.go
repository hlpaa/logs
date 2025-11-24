package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const errorLimit = 5
const interval = time.Minute

type ErrorPeak struct {
	UserID string `json:"user_id"`
	Count int `json:"count"`
	Interval time.Duration `json:"interval"`
}

func ErrorValidation(userLog Log){
	//implementar a verificação de pico de erros

	now := time.Now()
	var errorCount int

	// encontrar eventos de erro recentes
	for _, event := range userLog.Events {
		if strings.EqualFold(event.Severity, "ERROR") {
			eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
			if err != nil {
				continue
			}
			// verificar se o evento ocorreu dentro do intervalo de tempo
			if now.Sub(eventTime) <= interval {
				errorCount++
			}
		}
	}
	if( errorCount > errorLimit){
		report := ErrorPeak{
			UserID: userLog.UserID,
			Count: errorCount,
			Interval: interval,
		}

		errorPeak <- report
	}
}

func ErrorProducer(channel *amqp091.Channel) {
	for report := range errorPeak {
		// Cria a fila para user
		errorQueue, err := channel.QueueDeclare(
			"error." + report.UserID,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Print(err)
			continue
		}

		// Convert to JSON
		data, err := json.Marshal(report)
		if err != nil {
			log.Printf("Error marshaling JSON: %s", err)
		}

		err = channel.Publish(
			"",     			// exchange
			errorQueue.Name, 	// routing key
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
