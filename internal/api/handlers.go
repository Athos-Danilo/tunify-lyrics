package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// WorkerInterface define a interface que a API precisa do worker
type WorkerInterface interface {
	Trigger()
}

func registerHandlers(mux *http.ServeMux, worker WorkerInterface, logger *slog.Logger) {
	mux.HandleFunc("/health", healthHandler(logger))
	mux.HandleFunc("/trigger", triggerHandler(worker, logger))
}

func healthHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		
		_ = json.NewEncoder(w).Encode(response)
	}
}

func triggerHandler(worker WorkerInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		logger.Info("Disparo manual acionado via endpoint /trigger")
		
		// Acorda o Worker
		worker.Trigger()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		response := map[string]string{
			"message": "Sinal recebido. O processamento da fila foi iniciado em background, se já não estiver em andamento.",
		}
		
		_ = json.NewEncoder(w).Encode(response)
	}
}
