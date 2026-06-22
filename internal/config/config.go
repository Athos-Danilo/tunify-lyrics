// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig armazena as configurações essenciais do sistema
type AppConfig struct {
	MongoURI     string
	DatabaseName string
	CronInterval string
	Port         string
}

// Config contém a instância global de configuração
var Config AppConfig

// Load carrega as variáveis de ambiente do arquivo .env
// Utiliza biblioteca godotenv para ambiente local, e ignora se não existir (ex: Docker)
func Load() error {
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return fmt.Errorf("a variável de ambiente MONGO_URI é obrigatória")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		databaseName = "tunify" // Valor padrão de fallback
	}

	cronInterval := os.Getenv("CRON_INTERVAL")
	if cronInterval == "" {
		cronInterval = "@every 1m" // Padrão: rodar a cada 1 minuto
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	Config = AppConfig{
		MongoURI:     mongoURI,
		DatabaseName: databaseName,
		CronInterval: cronInterval,
		Port:         port,
	}

	return nil
}
