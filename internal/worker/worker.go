package worker

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/athosdanilo/tunify-letras/internal/config"
	"github.com/athosdanilo/tunify-letras/internal/lyrics"
	"github.com/athosdanilo/tunify-letras/internal/model"
	"github.com/athosdanilo/tunify-letras/internal/repository"
)

// LyricsWorker representa o worker assíncrono que busca letras
type LyricsWorker struct {
	cron            *cron.Cron
	letraRepo       *repository.LetraRepository
	cotaRepo        *repository.CotaRepository
	fallbackManager *lyrics.FallbackManager
	logger           *slog.Logger
	isRunning        bool
	cotaAtingidaData string
}

// NewLyricsWorker inicializa o Worker com o agendador Cron
func NewLyricsWorker(letraRepo *repository.LetraRepository, cotaRepo *repository.CotaRepository, fbManager *lyrics.FallbackManager, logger *slog.Logger) *LyricsWorker {
	c := cron.New()
	return &LyricsWorker{
		cron:            c,
		letraRepo:       letraRepo,
		cotaRepo:        cotaRepo,
		fallbackManager: fbManager,
		logger:          logger,
	}
}

// Start inicia o Cron Job
func (w *LyricsWorker) Start() error {
	_, err := w.cron.AddFunc(config.Config.CronInterval, w.processarFila)
	if err != nil {
		return err
	}
	
	_, err = w.cron.AddFunc("0 22 * * *", w.processarUpgrade)
	if err != nil {
		return err
	}

	w.cron.Start()
	w.logger.Info("Worker iniciado com sucesso", "intervalo", config.Config.CronInterval)
	w.logger.Info("Rotina noturna agendada para as 22h00")
	return nil
}

// Stop para o Cron Job
func (w *LyricsWorker) Stop() {
	w.cron.Stop()
	w.logger.Info("Worker parado")
}

// Trigger executa o ciclo do worker imediatamente em uma goroutine, ignorando o Cron
func (w *LyricsWorker) Trigger() {
	w.logger.Info("Trigger acionado!")
	go w.processarFila()
}

// processarFila é a função principal executada a cada "tick" do cron
func (w *LyricsWorker) processarFila() {
	if w.isRunning {
		w.logger.Warn("Worker pulou este ciclo porque o anterior ainda está em execução")
		return
	}
	w.isRunning = true
	defer func() { w.isRunning = false }()

	w.logger.Info("🔄 Iniciando ciclo de processamento da fila")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	dataHoje := time.Now().Format("2006-01-02")
	if w.cotaAtingidaData == dataHoje {
		w.logger.Info("💤 Cota diária já atingida hoje. Dormindo até amanhã...")
		return
	}

	cotaDiaria, err := w.cotaRepo.ObterCotaDoDia(ctx, dataHoje)
	if err != nil {
		w.logger.Error("Erro ao verificar cota diária", "erro", err)
		return
	}

	if cotaDiaria.ContagemGlobal >= config.Config.MaxDailyQuota {
		w.cotaAtingidaData = dataHoje
		w.logger.Warn("🛑 Cota diária global atingida. Entrando em retiro espiritual.", "cota", cotaDiaria.ContagemGlobal, "max", config.Config.MaxDailyQuota)
		return
	}

	// 1. Executar processamento de lote (ex: até 10 músicas por ciclo)
	batchSize := 10
	for i := 0; i < batchSize; i++ {
		// Trava a letra de forma atômica
		letra, err := w.letraRepo.BuscarMusicaPendente(ctx)
		if err != nil {
			if errors.Is(err, repository.ErrNoPendingLyrics) {
				if i == 0 {
					w.logger.Info("💤 Fila vazia. Nada a processar.")
				} else {
					w.logger.Info("💤 Não há mais músicas pendentes na fila.")
				}
				break
			}
			w.logger.Error("Erro ao buscar/travar música", "erro", err)
			continue
		}

		w.logger.Info("🎵 Processando música", "titulo", letra.Titulo, "artista", letra.Artista)

		status := model.StatusConcluido
		conteudo := ""
		sincronizada := false
		fonte := ""

		if strings.TrimSpace(letra.Titulo) == "" || strings.TrimSpace(letra.Artista) == "" {
			w.logger.Warn("Título ou artista vazios, ignorando busca", "titulo", letra.Titulo, "artista", letra.Artista)
			status = model.StatusNaoEncontrada
		} else {
			// Buscar letra nos provedores
			res, err := w.fallbackManager.FetchLyrics(ctx, letra.Artista, letra.Titulo)
			
			if err != nil {
				if errors.Is(err, lyrics.ErrNaoEncontrada) {
					status = model.StatusNaoEncontrada
					w.logger.Info("❌ Letra não encontrada em nenhum provedor")
				} else {
					// Retiro Espiritual de emergência: se tomamos 429 ou timeouts graves de infra, abortamos o lote todo.
					w.logger.Error("🚨 Erro de comunicação com o provedor. Abortando lote.", "erro", err)
					
					// Reverte o status para PENDENTE (fallback de segurança) para processar na próxima
					_ = w.letraRepo.AtualizarStatusMusica(ctx, letra.ID, model.StatusPendente, "", false, "", letra.TentativasProcessamento)
					return 
				}
			} else {
				conteudo = res.Letra
				sincronizada = res.Sincronizada
				fonte = res.Fonte
			}
		}

		// Salvar o resultado (incrementa tentativa de processamento ao terminar um ciclo do funil)
		err = w.letraRepo.AtualizarStatusMusica(ctx, letra.ID, status, conteudo, sincronizada, fonte, letra.TentativasProcessamento+1)
		if err != nil {
			w.logger.Error("Erro ao salvar resultado da música", "erro", err)
		}

		// Incrementar a cota global
		err = w.cotaRepo.IncrementarCota(ctx, dataHoje)
		if err != nil {
			w.logger.Error("Erro ao incrementar cota", "erro", err)
		}

		// Atualiza a memória local da cota global para saber se deve parar no meio do lote
		cotaDiaria.ContagemGlobal++
		if cotaDiaria.ContagemGlobal >= config.Config.MaxDailyQuota {
			w.cotaAtingidaData = dataHoje
			w.logger.Warn("Cota diária global atingida durante o lote. Abortando restante.")
			return
		}

		// JITTER / PAUSA HUMANA: Espera 5 segundos antes de avançar pra próxima música
		w.logger.Debug("⏳ Aplicando Jitter de 5s...")
		time.Sleep(5 * time.Second)
	}
}

// processarUpgrade tenta atualizar letras que não possuem sincronia utilizando uma cota paralela de 50 requests
func (w *LyricsWorker) processarUpgrade() {
	if w.isRunning {
		w.logger.Warn("⚠️ Worker noturno pulou este ciclo porque outro processamento está em execução")
		return
	}
	w.isRunning = true
	defer func() { w.isRunning = false }()

	w.logger.Info("🌙 Iniciando rotina noturna de upgrade de letras")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	dataHoje := time.Now().Format("2006-01-02")
	cotaDiaria, err := w.cotaRepo.ObterCotaDoDia(ctx, dataHoje)
	if err != nil {
		w.logger.Error("Erro ao verificar cota para upgrade", "erro", err)
		return
	}

	if cotaDiaria.ContagemNoturna >= 50 {
		w.logger.Info("Cota noturna (50 buscas) já foi atingida hoje. Dormindo...")
		return
	}

	limiteNoturno := 50 - cotaDiaria.ContagemNoturna
	
	// Executar em um único lote noturno
	for i := 0; i < limiteNoturno; i++ {
		letra, err := w.letraRepo.BuscarMusicaParaUpgrade(ctx)
		if err != nil {
			if errors.Is(err, repository.ErrNoPendingLyrics) {
				w.logger.Info("💤 Rotina noturna não encontrou mais nenhuma letra para dar upgrade. Finalizando ciclo.")
				break
			}
			w.logger.Error("🚨 Erro ao buscar música para upgrade", "erro", err)
			continue
		}

		w.logger.Info("✨ Tentando upgrade de música", "titulo", letra.Titulo, "artista", letra.Artista, "tentativa_atual", letra.TentativasProcessamento)

		// Busca forçada na API do fallbackManager
		res, err := w.fallbackManager.FetchLyrics(ctx, letra.Artista, letra.Titulo)
		
		status := model.StatusConcluido
		novaTentativa := letra.TentativasProcessamento + 1

		if err != nil || (res != nil && !res.Sincronizada) {
			// Se der erro ou se achar de novo uma não sincronizada, a gente só incrementa as tentativas
			// e mantém os dados velhos, exceto que o AtualizarStatusMusica requer os dados, 
			// então passamos os que ele já tinha antes (que estavam na struct)
			w.logger.Info("❌ Upgrade falhou, letra continua sem sincronia.", "titulo", letra.Titulo)
			err = w.letraRepo.AtualizarStatusMusica(ctx, letra.ID, status, letra.Conteudo, letra.Sincronizada, letra.Fonte, novaTentativa)
		} else {
			// Sucesso no upgrade
			w.logger.Info("✅ Upgrade realizado com sucesso! Letra sincronizada encontrada.", "titulo", letra.Titulo)
			err = w.letraRepo.AtualizarStatusMusica(ctx, letra.ID, status, res.Letra, res.Sincronizada, res.Fonte, novaTentativa)
		}

		if err != nil {
			w.logger.Error("🚨 Erro ao salvar upgrade da música", "erro", err)
		}

		err = w.cotaRepo.IncrementarCotaNoturna(ctx, dataHoje)
		if err != nil {
			w.logger.Error("🚨 Erro ao incrementar cota noturna", "erro", err)
		}

		w.logger.Debug("⏳ Aplicando Jitter noturno de 5s...")
		time.Sleep(5 * time.Second)
	}
	w.logger.Info("🏁 Rotina noturna de upgrade finalizada com sucesso.")
}
