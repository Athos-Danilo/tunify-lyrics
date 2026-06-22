package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ==========================================
// 🥇 NÍVEL OURO: LRCLIB (API REST)
// ==========================================

// RespostaLRCLIB mapeia o JSON que a API da LRCLIB nos devolve
type RespostaLRCLIB struct {
	ID           int    `json:"id"`
	TrackName    string `json:"trackName"`
	ArtistName   string `json:"artistName"`
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

func BuscarLRCLIB(artista, musica string) (texto string, sincronizada bool, sucesso bool) {
	// A LRCLIB é uma API amigável, então só precisamos montar a URL e bater nela
	query := url.Values{}
	query.Add("track_name", musica)
	query.Add("artist_name", artista)

	urlApi := "https://lrclib.net/api/search?" + query.Encode()

	// Adicionamos um Timeout para o robô não ficar travado se a API deles cair
	cliente := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", urlApi, nil)

	// É uma boa prática avisar quem somos no Header (User-Agent)
	req.Header.Set("User-Agent", "Tunify Letras Worker (Golang)")

	resp, err := cliente.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", false, false
	}
	defer resp.Body.Close()

	// A LRCLIB devolve uma lista (array) de resultados. Vamos pegar o primeiro.
	var resultados []RespostaLRCLIB
	if err := json.NewDecoder(resp.Body).Decode(&resultados); err != nil || len(resultados) == 0 {
		return "", false, false
	}

	primeiroResultado := resultados[0]

	// Se tiver letra com os milissegundos (karaokê), damos prioridade a ela!
	if primeiroResultado.SyncedLyrics != "" {
		return primeiroResultado.SyncedLyrics, true, true
	}

	// Se não tiver sincronizada, pegamos o texto puro pelo menos
	if primeiroResultado.PlainLyrics != "" {
		return primeiroResultado.PlainLyrics, false, true
	}

	return "", false, false
}

// ==========================================
// 🥈 NÍVEL PRATA: LETRAS.MUS.BR (Web Scraping)
// ==========================================

func BuscarLetrasMus(artista, musica string) (texto string, sucesso bool) {
	// O site do letras.mus.br usa URLs padronizadas: /nome-do-artista/nome-da-musica/
	// Precisamos transformar "Coldplay" em "coldplay" e espaços em hifens
	artistaSlug := strings.ToLower(strings.ReplaceAll(artista, " ", "-"))
	musicaSlug := strings.ToLower(strings.ReplaceAll(musica, " ", "-"))

	urlSite := fmt.Sprintf("https://www.letras.mus.br/%s/%s/", artistaSlug, musicaSlug)

	cliente := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", urlSite, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)") // Disfarce de navegador

	resp, err := cliente.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", false
	}
	defer resp.Body.Close()

	// Aqui usamos o bisturi do goquery para ler o HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", false
	}

	// O Letras.mus costuma guardar a letra dentro da div com a classe "lyric-original"
	var letraBuilder strings.Builder
	doc.Find(".lyric-original p").Each(func(i int, s *goquery.Selection) {
		// Substitui as tags <br> por quebras de linha reais do Go (\n)
		htmlLetra, _ := s.Html()
		textoLimpo := strings.ReplaceAll(htmlLetra, "<br/>", "\n")
		textoLimpo = strings.ReplaceAll(textoLimpo, "<br>", "\n")

		letraBuilder.WriteString(textoLimpo + "\n\n")
	})

	letraFinal := strings.TrimSpace(letraBuilder.String())
	if letraFinal == "" {
		return "", false
	}

	return letraFinal, true
}

// ==========================================
// 🥉 NÍVEL BRONZE: GENIUS (Web Scraping)
// ==========================================

func BuscarGenius(artista, musica string) (texto string, sucesso bool) {
	// A lógica do Genius segue o mesmo princípio de limpar os nomes para montar a URL
	artistaSlug := strings.Title(strings.ToLower(artista))
	artistaSlug = strings.ReplaceAll(artistaSlug, " ", "-")

	musicaSlug := strings.ToLower(strings.ReplaceAll(musica, " ", "-"))

	urlSite := fmt.Sprintf("https://genius.com/%s-%s-lyrics", artistaSlug, musicaSlug)

	cliente := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", urlSite, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := cliente.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", false
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", false
	}

	// O Genius usa a classe "Lyrics__Root" para envelopar os versos
	textoLetra := doc.Find("[class*='Lyrics__Root']").Text()
	textoLetra = strings.TrimSpace(textoLetra)

	if textoLetra == "" {
		return "", false
	}

	return textoLetra, true
}
