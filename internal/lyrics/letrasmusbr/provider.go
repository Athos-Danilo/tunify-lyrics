package letrasmusbr

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/athosdanilo/tunify-letras/internal/lyrics"
)

// Provider implementa LyricsProvider para o Letras.mus.br (Scraping)
type Provider struct {
	client *http.Client
}

// NewProvider cria um novo provider do Letras.mus.br
func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Name retorna o nome do provedor
func (p *Provider) Name() string {
	return "Letras.mus.br"
}

// Fetch busca a letra fazendo scraping do Letras.mus.br
func (p *Provider) Fetch(ctx context.Context, artista, titulo string) (*lyrics.Result, error) {
	artistaSlug := NormalizeForURL(artista)
	tituloSlug := NormalizeForURL(titulo)

	// URL alvo: https://www.letras.mus.br/{artistaSlug}/{tituloSlug}/
	reqURL := fmt.Sprintf("https://www.letras.mus.br/%s/%s/", artistaSlug, tituloSlug)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição Letras.mus.br: %w", err)
	}

	// User-Agent simulando um navegador real é crucial para scrapers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro HTTP ao acessar %s: %w", reqURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, lyrics.ErrNaoEncontrada
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Letras.mus.br retornou status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler HTML do Letras.mus.br: %w", err)
	}

	// A classe '.lyric-original' é padrão no Letras.mus.br atual
	letraDiv := doc.Find(".lyric-original")
	if letraDiv.Length() == 0 {
		return nil, lyrics.ErrNaoEncontrada
	}

	// Como o HTML usa tags <br> ou <p> para separar versos,
	// podemos usar .Text() mas pode ficar sem as quebras.
	// O goquery .Text() concatena tudo. Vamos extrair usando replace nas tags
	htmlContent, err := letraDiv.Html()
	if err != nil {
		return nil, fmt.Errorf("erro ao extrair HTML da div de letra: %w", err)
	}

	letraLimpa := limpaHTML(htmlContent)
	if letraLimpa == "" {
		return nil, lyrics.ErrNaoEncontrada
	}

	return &lyrics.Result{
		Letra:        letraLimpa,
		Sincronizada: false, // O letras.mus.br não fornece timestamp
	}, nil
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

// NormalizeForURL converte um nome em formato amigável para URL do letras.mus.br
// Ex: "Red Hot Chili Peppers" -> "red-hot-chili-peppers"
// "Beyoncé" -> "beyonce"
func NormalizeForURL(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = removeAccents(s)
	// Substituir caracteres não alfanuméricos por hífens
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	// Remover hífens duplicados ou nas extremidades
	s = strings.Trim(s, "-")
	return s
}

// limpaHTML converte <br> e <p> para \n e remove outras tags
func limpaHTML(html string) string {
	// Trocar tags que indicam quebra de linha por newline real
	brRegex := regexp.MustCompile(`(?i)<br\s*/?>|<\/p>\s*<p>`)
	text := brRegex.ReplaceAllString(html, "\n")
	
	// Remover qualquer outra tag HTML
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text = tagRegex.ReplaceAllString(text, "")
	
	// Remover quebras extras
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")

	return text
}

func removeAccents(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)
	return replacer.Replace(s)
}
