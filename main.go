package main

import (
	"context"       // Gerencia contexto e timeout de operações.
	"encoding/json" // Manipulação de JSON.
	"fmt"           // Formatação de strings para output.
	"log"           // Logging de erros e mensagens.
	"net/http"      // Servidor HTTP.
	"strings"       // Manipulação de strings.
	"time"          // Operações com tempo e duração.

	// Importamos nossos próprios pacotes.
	"tunify-lyrics-api/db"
	"tunify-lyrics-api/models"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Carrega as variáveis de ambiente do .env para a memória.
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Arquivo .env não encontrado. Usando variáveis do sistema.")
	}

	// Chama a função que criada no connection.go.
	db.ConectarMongoDB()

	// ROTA DE HEALTH CHECK:Todo microserviço profissional tem uma rota raiz ("/") que serve só para
	// ferramentas de monitoramento saberem se o servidor está vivo e respirando.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":      "Tunify Lyrics - API rodando lisa! ",
			"arquitetura": "Assíncrona (API + Fila no MongoDB)",
		})
	})

	// A ROTA PRINCIPAL (/lyrics): É aqui que onde o Angular (player.service.ts) vai bater.
	http.HandleFunc("/lyrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Libera o CORS.
		w.Header().Set("Content-Type", "application/json")

		// Extrai os parâmetros da URL (ex: ?artist=Coldplay&track=Yellow)
		// strings.TrimSpace remove espaços vazios que possam vir por acidente.
		artist := strings.TrimSpace(r.URL.Query().Get("artist"))
		track := strings.TrimSpace(r.URL.Query().Get("track"))

		// Validação de segurança: Não deixa o Angular pesquisar fantasmas.
		if artist == "" || track == "" {
			w.WriteHeader(http.StatusBadRequest) // Erro 400 (O cliente mandou errado).
			json.NewEncoder(w).Encode(map[string]string{"error": "Faltam os parâmetros artist ou track"})
			return
		}

		// Cria um contexto de 5 segundos para a busca no banco de dados.
		// Se o MongoDB Atlas engasgar, cancelamos a requisição para não travar o celular do usuário.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// bson.M (BSON Map): É assim que escrevemos filtros de busca (WHERE) no MongoDB pelo Go.
		// Estamos dizendo: "Procure um documento onde o nome_musica seja igual a 'track'..."
		filtro := bson.M{
			"nome_musica":  track,
			"nome_artista": artist,
		}

		// Cria uma variável vazia usando a nossa "forma" (a Struct que criada em letra.go).
		var letra models.Letra

		// Vai na coleção "letras", tenta achar UMA música com esse filtro, e se achar, "Decode" (injeta) os dados dentro da nossa variável 'letra'.
		err = db.LetrasCollection.FindOne(ctx, filtro).Decode(&letra)

		// ==========================================================
		// CASO 1: A MÚSICA NUNCA FOI TOCADA (Não existe no banco)
		// ==========================================================
		if err == mongo.ErrNoDocuments {
			// Preparamos o documento novinho em folha.
			novaLetra := models.Letra{
				NomeMusica:              track,
				NomeArtista:             artist,
				Status:                  "PENDENTE", // O semáforo amarelo pro Robô.
				TextoLetra:              nil,
				Sincronizada:            false,
				FonteLetra:              nil,
				TentativasProcessamento: 0,
				CriadoEm:                time.Now(),
				AtualizadoEm:            time.Now(),
			}

			// Inserimos a música na fila para o robô trabalhar em background.
			_, insertErr := db.LetrasCollection.InsertOne(ctx, novaLetra)
			if insertErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao agendar busca de letra"})
				return
			}

			// ARQUITETURA DE FILA:
			// Em vez de retornar 200 (OK), retornamos 202 (Accepted).
			// Isso significa: "Recebi seu pedido, aceitei ele, mas ainda não terminei de processar".
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "PENDENTE",
				"message": "Letra indisponível. Nosso robô estão processando esta música para as próximas execuções!",
			})
			return

		} else if err != nil { // Se deu algum outro erro bizarro no banco (caiu a internet, etc).
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro interno ao consultar o banco de dados"})
			return
		}

		// ==========================================================================
		// CASO 2: A MÚSICA JÁ EXISTE NO BANCO (Vamos ver em qual status ela está)
		// ==========================================================================
		switch letra.Status {
		case "CONCLUIDO":
			// Se o robô já fez o trabalho duro, entregamos a letra limpinha e em milissegundos.
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"artist":       letra.NomeArtista,
				"track":        letra.NomeMusica,
				"lyrics":       *letra.TextoLetra, // O '*' puxa o texto de dentro do ponteiro.
				"sincronizada": letra.Sincronizada,
			})

		case "PENDENTE", "PROCESSANDO":
			// O usuário ansioso pediu de novo, mas o robô ainda está trabalhando.
			w.WriteHeader(http.StatusAccepted) // 202 Accepted
			json.NewEncoder(w).Encode(map[string]string{
				"status":  letra.Status,
				"message": "Esta letra está na fila de processamento do robô. Tente novamente em breve!",
			})

		case "NAO_ENCONTRADA":
			// O robô já varreu o Genius, Letras.mus, LRCLIB e confirmou que essa música não tem letra.
			w.WriteHeader(http.StatusNotFound) // 404 Not Found.
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "NAO_ENCONTRADA",
				"message": "Letra indisponível para esta faixa.",
			})
		}
	})

	// Liga o Servidor.
	porta := ":8080"
	fmt.Printf("Tunify Letras API no ar! Escutando na porta %s...\n", porta)
	log.Fatal(http.ListenAndServe(porta, nil))
}
