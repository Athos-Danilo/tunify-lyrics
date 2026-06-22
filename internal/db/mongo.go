// internal/db/mongo.go
package db

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/athosdanilo/tunify-letras/internal/config"
	"github.com/athosdanilo/tunify-letras/internal/logger"
)

var (
	clientInstance *mongo.Client
	dbInstance     *mongo.Database
	clientError    error
	mongoOnce      sync.Once
)

// GetDatabase retorna a instância Singleton do banco de dados MongoDB.
// Utiliza sync.Once para garantir que a conexão seja aberta apenas uma vez.
func GetDatabase() (*mongo.Database, error) {
	mongoOnce.Do(func() {
		logger.Log.Info("Iniciando conexão com o MongoDB...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Configura o Connection Pool para suportar alta concorrência sem estourar limites
		clientOptions := options.Client().
			ApplyURI(config.Config.MongoURI).
			SetMaxPoolSize(100). 
			SetMinPoolSize(10)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			logger.Log.Error("Erro ao conectar no MongoDB", "error", err)
			clientError = err
			return
		}

		// Valida a conexão de fato com um ping
		err = client.Ping(ctx, nil)
		if err != nil {
			logger.Log.Error("Erro ao realizar ping no MongoDB", "error", err)
			clientError = err
			return
		}

		clientInstance = client
		dbInstance = client.Database(config.Config.DatabaseName)

		logger.Log.Info("Conexão com MongoDB estabelecida com sucesso!")
		
		// Inicializa os índices em background para não travar a subida da aplicação
		go ensureIndexes()
	})

	return dbInstance, clientError
}

// ensureIndexes cria os índices necessários na coleção letras para garantir performance máxima.
func ensureIndexes() {
	if dbInstance == nil {
		return
	}

	collection := dbInstance.Collection("letras")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Índice composto focado na busca da fila do Worker: ordenado por status e tempo
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "criado_em", Value: 1},
		},
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Log.Error("Erro ao criar índice no MongoDB", "error", err)
		return
	}

	logger.Log.Info("Índices do MongoDB verificados e criados com sucesso.")
}

// Disconnect encerra a conexão com o banco de dados de maneira graciosa.
func Disconnect(ctx context.Context) error {
	if clientInstance != nil {
		return clientInstance.Disconnect(ctx)
	}
	return nil
}
