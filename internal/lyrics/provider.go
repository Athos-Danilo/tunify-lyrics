package lyrics

import (
	"context"
	"errors"
)

// Result contém a letra encontrada e uma flag indicando se ela possui sincronia
type Result struct {
	Letra       string
	Sincronizada bool
}

// LyricsProvider define o contrato para os motores de busca de letras
type LyricsProvider interface {
	// Name retorna o nome do provedor (útil para logs)
	Name() string
	// Fetch busca a letra de uma música dado o artista e o título
	Fetch(ctx context.Context, artista string, titulo string) (*Result, error)
}

// Erros comuns
var (
	ErrNaoEncontrada = errors.New("letra não encontrada")
)
