// cmd/server/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/athosdanilo/tunify-letras/internal/config"
	"github.com/athosdanilo/tunify-letras/internal/db"
	"github.com/athosdanilo/tunify-letras/internal/logger"
)

func main() {
	// Inicializa o logger principal
	logger.Init()
	logger.Log.Info("Iniciando Tunify Letras Microservice...")

	// Carrega as configurações do ambiente (.env)
	if err := config.Load(); err != nil {
		logger.Log.Error("Falha ao carregar configurações", "error", err)
		os.Exit(1)
	}

	// Tenta obter a instância do banco de dados MongoDB (aciona o Singleton)
	_, err := db.GetDatabase()
	if err != nil {
		logger.Log.Error("Falha crítica de banco de dados, parando o serviço", "error", err)
		os.Exit(1)
	}

	logger.Log.Info("Setup Base Concluído com Sucesso! (Épico 1)")
	logger.Log.Info("Aguardando interrupção do sistema (Ctrl+C)...")

	// Prepara um canal para escutar sinais do sistema operacional (ex: SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Bloqueia a execução principal até receber um sinal no canal
	<-quit

	// Inicia o processo de desligamento gracioso (Graceful Shutdown)
	logger.Log.Info("Sinal de interrupção recebido. Desligando o sistema de forma graciosa...")
	
	// Dá até 5 segundos para limpar tudo antes de forçar a saída
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Fecha a conexão com o MongoDB
	if err := db.Disconnect(ctx); err != nil {
		logger.Log.Error("Erro ao desconectar do MongoDB", "error", err)
	}
	
	logger.Log.Info("Sistema encerrado com segurança. Até a próxima!")
}
