// cmd/server/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/athosdanilo/tunify-letras/internal/api"
	"github.com/athosdanilo/tunify-letras/internal/config"
	"github.com/athosdanilo/tunify-letras/internal/db"
	"github.com/athosdanilo/tunify-letras/internal/logger"
	"github.com/athosdanilo/tunify-letras/internal/lyrics"
	"github.com/athosdanilo/tunify-letras/internal/lyrics/letrasmusbr"
	"github.com/athosdanilo/tunify-letras/internal/lyrics/lrclib"
	"github.com/athosdanilo/tunify-letras/internal/repository"
	"github.com/athosdanilo/tunify-letras/internal/worker"
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

	// Obtém a instância do banco de dados MongoDB (Singleton)
	database, err := db.GetDatabase()
	if err != nil {
		logger.Log.Error("Falha crítica de banco de dados, parando o serviço", "error", err)
		os.Exit(1)
	}

	// 1. Inicializa os Repositórios
	letraRepo := repository.NewLetraRepository(database)
	cotaRepo := repository.NewCotaRepository(database)

	// 2. Inicializa os Motores de Letras e o FallbackManager
	ouroProv := lrclib.NewProvider()
	prataProv := letrasmusbr.NewProvider()
	fallbackMgr := lyrics.NewFallbackManager(logger.Log, ouroProv, prataProv)

	// 3. Inicializa o Motor Assíncrono (Worker)
	lyricsWorker := worker.NewLyricsWorker(letraRepo, cotaRepo, fallbackMgr, logger.Log)
	if err := lyricsWorker.Start(); err != nil {
		logger.Log.Error("Erro ao iniciar o Worker", "error", err)
		os.Exit(1)
	}

	// 4. Inicializa e sobe a API HTTP
	apiServer := api.NewServer(config.Config.Port, lyricsWorker, logger.Log)
	go func() {
		if err := apiServer.Start(); err != nil {
			logger.Log.Error("Servidor HTTP falhou", "error", err)
		}
	}()

	logger.Log.Info("Todas as engrenagens ativadas. Sistema rodando perfeitamente!")
	logger.Log.Info("Aguardando interrupção do sistema (Ctrl+C)...")

	// Prepara um canal para escutar sinais do sistema operacional (ex: SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Bloqueia a execução principal até receber um sinal no canal
	<-quit

	// Inicia o processo de desligamento gracioso (Graceful Shutdown)
	logger.Log.Info("Sinal de interrupção recebido. Desligando o sistema de forma graciosa...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Para o servidor HTTP
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Log.Error("Erro no shutdown da API HTTP", "error", err)
	}

	// Para o cron job do Worker
	lyricsWorker.Stop()

	// Fecha a conexão com o MongoDB
	if err := db.Disconnect(ctx); err != nil {
		logger.Log.Error("Erro ao desconectar do MongoDB", "error", err)
	}
	
	logger.Log.Info("Sistema encerrado com segurança. Até a próxima!")
}
