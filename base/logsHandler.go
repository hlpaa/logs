package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

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
	if fromStr != "" {
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
