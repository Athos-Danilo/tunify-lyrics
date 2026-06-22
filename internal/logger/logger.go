// internal/logger/logger.go
package logger

import (
	"log/slog"
	"os"
)

// Log é a instância global do logger estruturado (JSON)
var Log *slog.Logger

// Init inicializa o logger da aplicação
// Utiliza slog (nativo do Go) para gerar logs no formato JSON, ideal para Cloud/Observabilidade
func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	
	// Configura saída padrão para JSON
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler)
	
	// Substitui o logger global do pacote "log" para usar nosso novo formato
	slog.SetDefault(Log)
}
