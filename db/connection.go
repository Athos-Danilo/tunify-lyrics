package db

import (
	"context" // Gerencia contexto e timeout de operações.
	"fmt"     // Formatação de strings para output.
	"log"     // Logging de erros e mensagens.
	"os"      // Acesso a variáveis de ambiente.
	"time"    // Operações com tempo e duração.

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var LetrasCollection *mongo.Collection

func ConectarMongoDB() {
	// Puxa a string de conexão (URL + Senha) do arquivo .env ou da nuvem.
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatal("ERRO FATAL: Variável MONGO_URI não encontrada no ambiente!")
	}

	// Se o banco não responder em 10 segundos, o Go encerra a conexão.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer garante que o timeout será cancelado ao fim da função
	defer cancel()

	// Prepara as configurações de conexão com a URL.
	clientOptions := options.Client().ApplyURI(uri)

	// Inicia a criação do cliente. Importante: isso ainda não conecta fisicamente no banco!
	// Apenas prepara a estrutura na memória usando aquele limite de 10 segundos (ctx).
	client, err := mongo.Connect(ctx, clientOptions)

	// Verifica se houve erro na preparação do cliente
	if err != nil {
		log.Fatal("Erro fatal ao preparar cliente do MongoDB: ", err)
	}

	// Testa a conectividade real com o servidor MongoDB (handshake).
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Erro ao dar Ping no MongoDB (Sua rede pode estar bloqueando): ", err)
	}

	fmt.Println("Conectado ao MongoDB Atlas com segurança via .env!")

	// Armazena o cliente global para uso em outras funções.
	Client = client

	// Obtém referência da coleção de letras (será usada para CRUD).
	LetrasCollection = client.Database("tunify").Collection("letras")
}
