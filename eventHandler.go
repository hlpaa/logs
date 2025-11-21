package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// POST /event
type EventRequest struct {
	UserID    string `json:"user_id"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

const errorLimit = 5
const interval = time.Minute

func checkErrorPeak(userLog Log){
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
	fmt.Printf("Usuário %s tem %d erros nos últimos %s\n", userLog.UserID, errorCount, interval)
	if( errorCount > errorLimit){
		fmt.Printf("Alerta: Usuário %s excedeu o limite de erros com %d erros nos últimos %s\n", userLog.UserID, errorCount, interval)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	// POST /event
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var EventReq EventRequest
	err := json.NewDecoder(r.Body).Decode(&EventReq)
	if err != nil {
		http.Error(w, "Erro ao analisar o JSON do log", http.StatusBadRequest)
		return
	}

	if EventReq.UserID == "" || EventReq.Timestamp == "" || EventReq.Severity == "" || EventReq.Message == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Evento registrado com sucesso")
	checkErrorPeak(*currentLog)
}
