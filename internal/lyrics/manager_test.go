package lyrics_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/athosdanilo/tunify-letras/internal/lyrics"
)

// MockProvider é um mock simples para testes
type MockProvider struct {
	NameFunc  func() string
	FetchFunc func(ctx context.Context, artista, titulo string) (*lyrics.Result, error)
}

func (m *MockProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "Mock"
}

func (m *MockProvider) Fetch(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(ctx, artista, titulo)
	}
	return nil, lyrics.ErrNaoEncontrada
}

func TestFallbackManager_FetchLyrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("Ouro encontra a letra", func(t *testing.T) {
		ouro := &MockProvider{
			NameFunc: func() string { return "Ouro" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				return &lyrics.Result{Letra: "letra ouro", Sincronizada: true}, nil
			},
		}
		prata := &MockProvider{
			NameFunc: func() string { return "Prata" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				t.Fatal("Prata não deveria ser chamado")
				return nil, nil
			},
		}

		manager := lyrics.NewFallbackManager(logger, ouro, prata)
		res, err := manager.FetchLyrics(context.Background(), "Artista", "Musica")
		
		if err != nil {
			t.Errorf("Erro não esperado: %v", err)
		}
		if res.Letra != "letra ouro" {
			t.Errorf("Esperado 'letra ouro', obtido '%s'", res.Letra)
		}
	})

	t.Run("Ouro falha e Prata encontra", func(t *testing.T) {
		ouro := &MockProvider{
			NameFunc: func() string { return "Ouro" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				return nil, lyrics.ErrNaoEncontrada
			},
		}
		prata := &MockProvider{
			NameFunc: func() string { return "Prata" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				return &lyrics.Result{Letra: "letra prata", Sincronizada: false}, nil
			},
		}

		manager := lyrics.NewFallbackManager(logger, ouro, prata)
		res, err := manager.FetchLyrics(context.Background(), "Artista", "Musica")
		
		if err != nil {
			t.Errorf("Erro não esperado: %v", err)
		}
		if res.Letra != "letra prata" {
			t.Errorf("Esperado 'letra prata', obtido '%s'", res.Letra)
		}
	})

	t.Run("Nenhum encontra", func(t *testing.T) {
		ouro := &MockProvider{
			NameFunc: func() string { return "Ouro" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				return nil, lyrics.ErrNaoEncontrada
			},
		}
		prata := &MockProvider{
			NameFunc: func() string { return "Prata" },
			FetchFunc: func(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
				return nil, errors.New("timeout interno do prata")
			},
		}

		manager := lyrics.NewFallbackManager(logger, ouro, prata)
		_, err := manager.FetchLyrics(context.Background(), "Artista", "Musica")
		
		if !errors.Is(err, lyrics.ErrNaoEncontrada) {
			t.Errorf("Esperado erro %v, obtido %v", lyrics.ErrNaoEncontrada, err)
		}
	})
}
