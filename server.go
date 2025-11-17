package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const port = ":8080"

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Severity	string `json:"severity"`
	Message		string `json:"message"`
}

var logs []LogEntry

func logsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch r.Method {

	// Armazenar logs enviados via POST
	case http.MethodPost:
		addLogsHandler(w, r)
		return


	// Consultas dos logs via GET 
	case http.MethodGet:
		getLogsHandler(w, r)
		return

	default:
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
		return
	}
}

func addLogsHandler(w http.ResponseWriter, r *http.Request) {
		var entry LogEntry
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "Erro ao analisar o JSON do log", http.StatusBadRequest)
			return
		}
		logs = append(logs, entry)
		w.WriteHeader(http.StatusCreated)	
		fmt.Fprintf(w, "Log recebido e armazenado com sucesso.\n")
}

func getLogsHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
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

		//criação do array de logs filtrados
		filtered := make([]LogEntry, 0, len(logs))
		for _,le := range logs {
			// se tiver filtro de severidade na req e não bater com o valor que procuramos no filtro, pule
			if severityFilter != "" && !strings.EqualFold(le.Severity, severityFilter){
				continue
			}

			// se tiver filtro de intervalo de tempo na req
			if fromStr != "" || toStr != "" {
				if le.Timestamp == ""{
					continue
				}
				t, perr := time.Parse(time.RFC3339, le.Timestamp)
				if perr != nil {
						continue
				}
				// se não bater com o intervalo de tempo que procuramos no filtro, pule
				if !fromTime.IsZero() && t.Before(fromTime) {
						continue
				}
				if !toTime.IsZero() && t.After(toTime) {
						continue
				}
				filtered = append(filtered, le)
			}
		}

		if err := json.NewEncoder(w).Encode(logs); err != nil {
			http.Error(w, "Erro ao codificar os logs em JSON", http.StatusInternalServerError)
			return
		}
}

func main() {
	
	http.HandleFunc("/logs", logsHandler)


	fmt.Printf("Servidor iniciado e escutando em http://localhost%s\n", port)
	
	err := http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %s\n", err)
	}
}