package main

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type EventRequest struct {
	UserID    string `json:"user_id"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

func LoggingHandler(message *amqp091.Delivery) {
	var EventReq EventRequest
	err := json.Unmarshal(message.Body, &EventReq)
	if err != nil {
		log.Print(err)
		return
	}

	if EventReq.UserID == "" || EventReq.Timestamp == "" || EventReq.Severity == "" || EventReq.Message == "" {
		log.Print("Todos os campos são obrigatórios")
		return
	}

	// encontrar o log do usuário ou criar um novo
	found := false
	var currentLog *Log
	for i, log := range logs {
		if log.UserID == EventReq.UserID {
			// adicionar o evento ao log existente
			logs[i].Events = append(logs[i].Events, Event{
				Timestamp: EventReq.Timestamp,
				Severity:  EventReq.Severity,
				Message:   EventReq.Message,
			})
			currentLog = &logs[i]
			found = true
			break
		}
	}
	if !found {
		// criar um novo log para o usuário
		newLog := Log{
			UserID: EventReq.UserID,
			Events: []Event{
				{
					Timestamp: EventReq.Timestamp,
					Severity:  EventReq.Severity,
					Message:   EventReq.Message,
				},
			},
		}
		logs = append(logs, newLog)
		currentLog = &logs[len(logs)-1]
	}

	if (EventReq.Severity == "ERROR") {
		ErrorValidation(*currentLog)
	}
}

func LoggingConsumer(conn *amqp091.Connection) {
	// Cria um canal
	logChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer logChannel.Close()

	// Cria a fila
	loggingQueue, err := logChannel.QueueDeclare(
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

	// Consome mensagens
	messages, err := logChannel.Consume(
		loggingQueue.Name, // fila
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
		for delivered := range messages {
			go LoggingHandler(&delivered)
		}
	}()

	log.Printf("Aguardando logs...")
	<-forever
}
