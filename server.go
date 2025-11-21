package main

import (
	"fmt"
	"net/http"
)

const port = ":8080"

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

func main() {
	
	http.HandleFunc("/event", eventHandler)
	http.HandleFunc("/logs/", logsHandler)


	fmt.Printf("Servidor iniciado e escutando em http://localhost%s\n", port)
	
	err := http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %s\n", err)
	}
}