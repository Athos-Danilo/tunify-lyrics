package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Server representa o servidor HTTP
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// NewServer inicializa um novo servidor HTTP e registra as rotas
func NewServer(port string, worker WorkerInterface, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// Registro das rotas
	registerHandlers(mux, worker, logger)

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}

// Start sobe o servidor. Ele bloqueia a goroutine até que ocorra um erro ou shutdown.
func (s *Server) Start() error {
	s.logger.Info("Servidor HTTP rodando", "porta", s.httpServer.Addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown desliga o servidor de forma graciosa.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
