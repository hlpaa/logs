package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type QueryRequest struct {
	UserID string  `json:"user_id"`
	Severity string `json:"severity"`
	From string `json:"from"`
	To string `json:"to"`
}

func QueryHandler(message *amqp091.Delivery) {
	var queryReq QueryRequest
	err := json.Unmarshal(message.Body, &queryReq)
	if err != nil {
		log.Print(err)
		return
	}

	if queryReq.UserID == "" {
		log.Print("O ID é obrigatório")
		return
	}

	//encontrar o log do usuário
	var userLog *Log
	for i, log := range logs {
		if log.UserID == queryReq.UserID {
			userLog = &logs[i]
			break
		}
	}

	if userLog == nil {
		log.Print("Log do usuário não encontrado")
		return
	}

	// extrair os filtros da query
	severityFilter := queryReq.Severity
	fromStr := queryReq.From
	toStr := queryReq.To

	var fromTime, toTime time.Time
	if fromStr != "" {
		fromTime, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			log.Print("Parâmetro 'from' inválido; use RFC3339")
			return
		}
	}
	if toStr != "" {
		toTime, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			log.Print("parâmetro 'to' inválido; use RFC3339")
			return
		}
	}

	// filtrar os eventos do userLog
	filteredEvents := make([]Event, 0)
	for _, event := range userLog.Events {

		//checa se tem filtro de severidade na req e se o evento bate com o filtro
		if severityFilter != "" && !strings.EqualFold(event.Severity, severityFilter) {
			continue
		}

		//checa se tem filtro de from e to na req e se o evento bate com os filtros
		if fromStr != "" || toStr != "" {
			if event.Timestamp == "" {
				continue
			}
			eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
			if err != nil {
				continue
			}
			if !fromTime.IsZero() && eventTime.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && eventTime.After(toTime) {
				continue
			}
		}
		filteredEvents = append(filteredEvents, event)
	}

	// criar um log com os eventos filtrados
	result := Log{
		UserID: userLog.UserID,
		Events: filteredEvents,
	}

	queryResult <- result
}

func QueryConsumer(conn *amqp091.Connection) {
	// Cria um canal
	queryChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer queryChannel.Close()

	// Cria a fila
	queryQueue, err := queryChannel.QueueDeclare(
		"event.query",
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
		for delivered := range messages {
			go QueryHandler(&delivered)
		}
	}()

	log.Print("Aguardando queries...")
	<-forever
}

func ResultProducer(channel *amqp091.Channel) {
	for query := range queryResult {
		// Cria a fila para user
		queryQueue, err := channel.QueueDeclare(
			"query." + query.UserID,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Print(err)
		}

		// Convert to JSON
		data, err := json.Marshal(query)
		if err != nil {
			log.Print("Error marshaling JSON:", err)
		}

		err = channel.Publish(
			"",     			// exchange
			queryQueue.Name, 	// routing key
			false,  			// mandatory
			false,  			// immediate
			amqp091.Publishing{
				Body: data,
				ContentType: "application/json", 	// Important!
				DeliveryMode: amqp091.Persistent, 	// Make message persistent
			})
		if err != nil {
			log.Fatal(err)
		}
	}
}
