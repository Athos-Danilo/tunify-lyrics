package lyrics

import (
	"context"
	"errors"
	"log/slog"
)

// FallbackManager gerencia a busca em cascata entre múltiplos provedores
type FallbackManager struct {
	providers []LyricsProvider
	logger    *slog.Logger
}

// NewFallbackManager cria uma nova instância do FallbackManager
func NewFallbackManager(logger *slog.Logger, providers ...LyricsProvider) *FallbackManager {
	return &FallbackManager{
		providers: providers,
		logger:    logger,
	}
}

// FetchLyrics tenta buscar a letra iterando pelos provedores na ordem definida
func (fm *FallbackManager) FetchLyrics(ctx context.Context, artista, titulo string) (*Result, error) {
	fm.logger.Info("Iniciando busca de letra", "artista", artista, "titulo", titulo)

	for _, provider := range fm.providers {
		fm.logger.Debug("Tentando provedor", "provedor", provider.Name())
		
		res, err := provider.Fetch(ctx, artista, titulo)
		if err == nil && res != nil {
			res.Fonte = provider.Name()
			fm.logger.Info("Letra encontrada", "provedor", provider.Name(), "sincronizada", res.Sincronizada)
			return res, nil
		}

		if ctx.Err() != nil {
			fm.logger.Error("Contexto global cancelado ou tempo excedido", "erro", ctx.Err())
			return nil, ctx.Err()
		}

		// Se não encontrou ou deu erro (ex: timeout interno do provedor), loga e tenta o próximo
		fm.logger.Warn("Falha no provedor, tentando próximo...", "provedor", provider.Name(), "erro", err)
	}

	fm.logger.Info("Nenhum provedor conseguiu encontrar a letra", "artista", artista, "titulo", titulo)
	return nil, ErrNaoEncontrada
}
