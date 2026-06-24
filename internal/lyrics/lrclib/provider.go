package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/athosdanilo/tunify-letras/internal/lyrics"
)

// Provider implementa LyricsProvider para a API pública do LRCLIB
type Provider struct {
	client *http.Client
}

// NewProvider cria um novo provider do LRCLIB
func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 10 * time.Second, // Timeout adequado para não travar
		},
	}
}

// Name retorna o nome do provedor
func (p *Provider) Name() string {
	return "LRCLIB"
}

// lrclibResponse espelha o formato JSON retornado pela API
type lrclibResponse struct {
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

// Fetch busca a letra na API do LRCLIB
func (p *Provider) Fetch(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
	// Construir URL: https://lrclib.net/api/get?artist_name={artist}&track_name={track}
	baseURL := "https://lrclib.net/api/get"
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear URL do LRCLIB: %w", err)
	}

	q := reqURL.Query()
	q.Set("artist_name", artista)
	q.Set("track_name", titulo)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição LRCLIB: %w", err)
	}
	
	// Adicionando um User-Agent por boas práticas e para evitar bloqueios bobos
	req.Header.Set("User-Agent", "Tunify-Lyrics-Worker/1.0 (github.com/athosdanilo/tunify-letras)")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar LRCLIB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, lyrics.ErrNaoEncontrada
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LRCLIB retornou status code não esperado: %d", resp.StatusCode)
	}

	var data lrclibResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON do LRCLIB: %w", err)
	}

	// Prioridade: Letra Sincronizada > Letra Plana
	if data.SyncedLyrics != "" {
		return &lyrics.Result{
			Letra:        data.SyncedLyrics,
			Sincronizada: true,
		}, nil
	}

	if data.PlainLyrics != "" {
		return &lyrics.Result{
			Letra:        data.PlainLyrics,
			Sincronizada: false,
		}, nil
	}

	return nil, lyrics.ErrNaoEncontrada
}
