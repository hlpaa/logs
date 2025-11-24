package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"

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

type QueryRequest struct {
	UserID string  `json:"user_id"`
	Severity string `json:"severity"`
	From string `json:"from"`
	To string `json:"to"`
}

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

	reader := bufio.NewReader(os.Stdin)
	log.Print("Enter User ID: ")

	userID, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
		return
	}

	go QueryConsumer(conn, userID)
	go AlertsConsumer(conn, userID)
	go LoggingProducer(conn, userID)

	queue, err := channel.QueueDeclare(
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

	log.Print("Press enter to make a query")
	for {
		severity, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			return
		}

		query := QueryRequest{
			UserID: userID,
			Severity: strings.TrimSpace(severity),
			From: "",
			To: "",
		}

		data, err := json.Marshal(query)
		if err != nil {
			log.Printf("Error marshaling JSON: %s", err)
		}

		err = channel.Publish(
			"",     			// exchange
			queue.Name, 		// routing key
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
