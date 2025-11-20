package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const port = ":8080"

type Event struct {
    Timestamp string `json:"timestamp"`
    Severity  string `json:"severity"`
    Message   string `json:"message"`
}

// POST /event
type EventRequest struct {
    UserID    string `json:"user_id"`
    Timestamp string `json:"timestamp"`
    Severity  string `json:"severity"`
    Message   string `json:"message"`
}

type Log struct {
    UserID string  `json:"user_id"`
    Events []Event `json:"events"`
}

var logs []Log

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

		if(EventReq.UserID == "" || EventReq.Timestamp == "" || EventReq.Severity == "" || EventReq.Message == ""){
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

func logsHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		

		// extrari o userID da URL
		path := r.URL.Path
		userID := strings.TrimPrefix(path, "/logs/")

		if userID == "" {
			http.Error(w, "userID é obrigatório na URL", http.StatusBadRequest)
			return
		}

		//encontrar o log do usuário 
		var userLog *Log
		for i, log := range logs {
			if log.UserID == userID {
				userLog = &logs[i]
				break
			}
		}

		if userLog == nil {
			http.Error(w, "Log do usuário não encontrado", http.StatusNotFound)
			return
		}

		q := r.URL.Query()
		// extrar os filtros da query
		severityFilter := q.Get("severity")
		fromStr := q.Get("from")
		toStr := q.Get("to")

		var fromTime, toTime time.Time
		var err error
		if fromStr != ""{
			fromTime, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				http.Error(w, "parâmetro 'from' inválido; use RFC3339", http.StatusBadRequest)
			}
		}
		if toStr != "" {
			toTime, err = time.Parse(time.RFC3339, toStr)
			if err != nil {
					http.Error(w, "parâmetro 'to' inválido; use RFC3339", http.StatusBadRequest)
					return
			}
		}

		// filtrar os eventos do userLog
		filteredEvents := make([]Event, 0, len(userLog.Events))
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

		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "Erro ao codificar os logs em JSON", http.StatusInternalServerError)
			return
		}
}

func main() {
	
	http.HandleFunc("/event", eventHandler)
	http.HandleFunc("/logs/", logsHandler)


	fmt.Printf("Servidor iniciado e escutando em http://localhost%s\n", port)
	
	err := http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %s\n", err)
	}
}