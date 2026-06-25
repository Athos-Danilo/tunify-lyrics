// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig armazena as configurações essenciais do sistema
type AppConfig struct {
	MongoURI         string
	DatabaseName     string
	CronInterval     string
	Port             string
	MaxDailyQuota    int
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
		cronInterval = "*/15 * * * *" // Padrão: rodar a cada 15 minutos alinhado ao relógio (ex: 05:00, 05:15...)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	maxDailyQuotaStr := os.Getenv("MAX_DAILY_QUOTA")
	maxDailyQuota := 100
	if maxDailyQuotaStr != "" {
		fmt.Sscanf(maxDailyQuotaStr, "%d", &maxDailyQuota)
	}

	Config = AppConfig{
		MongoURI:         mongoURI,
		DatabaseName:     databaseName,
		CronInterval:     cronInterval,
		Port:             port,
		MaxDailyQuota:    maxDailyQuota,
	}

	return nil
}
