package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// POST /event
type EventRequest struct {
	UserID    string `json:"user_id"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
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
	for i, log := range logs {
		if log.UserID == EventReq.UserID {
			// adicionar o evento ao log existente
			logs[i].Events = append(logs[i].Events, Event{
				Timestamp: EventReq.Timestamp,
				Severity:  EventReq.Severity,
				Message:   EventReq.Message,
			})
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
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Evento registrado com sucesso")
}
