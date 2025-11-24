package main

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type Event struct {
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type Log struct {
	UserID string  `json:"user_id"`
	Events []Event `json:"events"`
}

var logs []Log
var errorPeak chan ErrorPeak
var queryResult chan Log

func main() {
	// Conecta ao RabbitMQ
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
    	log.Fatal(err)
	}
	defer conn.Close()

	// Cria um canal
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	forever := make(chan bool)
	errorPeak = make(chan ErrorPeak, 10)
	queryResult = make(chan Log)

	go QueryConsumer(conn)
	go LoggingConsumer(conn)
	
	go ErrorProducer(channel)
	go ResultProducer(channel)

	<-forever
}
