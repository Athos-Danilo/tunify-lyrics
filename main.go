package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 🚨 1. FORMATA A URL: Transforma "Linkin Park" em "linkin-park"
func formatGeniusURL(artist, track string) string {
	// Tudo minúsculo
	artist = strings.ToLower(artist)
	track = strings.ToLower(track)

	// Troca espaços por hífens
	artist = strings.ReplaceAll(artist, " ", "-")
	track = strings.ReplaceAll(track, " ", "-")

	// Tira as aspas simples e caracteres que quebram o link
	artist = strings.ReplaceAll(artist, "'", "")
	track = strings.ReplaceAll(track, "'", "")

	// Monta o padrão do site: genius.com/Artista-musica-lyrics
	return fmt.Sprintf("https://genius.com/%s-%s-lyrics", artist, track)
}

// 🚨 2. O BISTURI DE SCRAPING: Entra no site e corta a letra
func scrapeGeniusLyrics(artist, track string) (string, error) {
	url := formatGeniusURL(artist, track)
	fmt.Println("[TUNIFY LOG] Varrendo a página:", url)

	// Faz a visita na página igual a um navegador
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("página não encontrada (Status %d)", res.StatusCode)
	}

	// Carrega o HTML inteiro para o Goquery ler
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var lyricsText string

	// O Genius guarda as letras dentro de containers com a tag data-lyrics-container="true"
	doc.Find("[data-lyrics-container='true']").Each(func(i int, s *goquery.Selection) {
		// Pegamos o HTML cru daquele pedaço
		htmlStr, _ := s.Html()

		// 🚨 O TRUQUE DE MESTRE: Transformar as tags <br> do HTML em quebra de linha real (\n)
		htmlStr = strings.ReplaceAll(htmlStr, "<br/>", "\n")
		htmlStr = strings.ReplaceAll(htmlStr, "<br>", "\n")

		// Limpamos qualquer outra tag (links embutidos) lendo o texto de novo
		cleanDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
		lyricsText += cleanDoc.Text() + "\n\n"
	})

	if lyricsText == "" {
		return "", fmt.Errorf("container de letras não encontrado")
	}

	// Tira espaços extras do começo e do fim
	return strings.TrimSpace(lyricsText), nil
}

// 🚨 3. O SERVIDOR: Recebe a chamada do Angular e devolve o JSON
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "Tunify Lyrics API rodando lisa! 🚀",
		})
	})

	http.HandleFunc("/lyrics", func(w http.ResponseWriter, r *http.Request) {
		// Liberando o CORS pro Angular não tomar bloqueio
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Pegando os dados da URL (?artist=...&track=...)
		artist := r.URL.Query().Get("artist")
		track := r.URL.Query().Get("track")

		if artist == "" || track == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Faltam os parâmetros artist ou track"})
			return
		}

		// Chama a nossa função de extração lá de cima
		lyrics, err := scrapeGeniusLyrics(artist, track)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Letra não encontrada",
				"details": err.Error(),
			})
			return
		}

		// Devolve a letra de verdade pro Angular!
		json.NewEncoder(w).Encode(map[string]string{
			"artist": artist,
			"track":  track,
			"lyrics": lyrics,
		})
	})

	porta := ":8080"
	fmt.Printf("📻 Tunify Lyrics API no ar! Escutando na porta %s...\n", porta)
	log.Fatal(http.ListenAndServe(porta, nil))
}
