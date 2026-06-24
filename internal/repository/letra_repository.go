package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/athosdanilo/tunify-letras/internal/model"
)

var ErrNoPendingLyrics = errors.New("nenhuma letra pendente encontrada")

// LetraRepository provê as operações de banco de dados para a coleção de letras
type LetraRepository struct {
	collection *mongo.Collection
}

// NewLetraRepository cria uma nova instância de LetraRepository
func NewLetraRepository(db *mongo.Database) *LetraRepository {
	return &LetraRepository{
		collection: db.Collection("letras"),
	}
}

// BuscarUsuariosComPendencias retorna uma lista de IDs únicos de usuários que possuem letras pendentes
func (r *LetraRepository) BuscarUsuariosComPendencias(ctx context.Context) ([]primitive.ObjectID, error) {
	filter := bson.M{"status": model.StatusPendente}
	
	resultados, err := r.collection.Distinct(ctx, "id_usuario", filter)
	if err != nil {
		return nil, err
	}

	var usuarios []primitive.ObjectID
	for _, res := range resultados {
		if id, ok := res.(primitive.ObjectID); ok {
			usuarios = append(usuarios, id)
		}
	}
	
	return usuarios, nil
}

// BuscarMusicaPendentePorUsuario busca a música PENDENTE mais antiga de um usuário e altera atômicamente seu status para PROCESSANDO.
// Garante ausência de Race Conditions pelo uso de FindOneAndUpdate.
func (r *LetraRepository) BuscarMusicaPendentePorUsuario(ctx context.Context, idUsuario primitive.ObjectID) (*model.Letra, error) {
	filter := bson.M{
		"status":     model.StatusPendente,
		"id_usuario": idUsuario,
	}
	
	update := bson.M{
		"$set": bson.M{
			"status":        model.StatusProcessando,
			"atualizado_em": time.Now(),
		},
	}

	// Ordena por criado_em ascendente (mais antigas primeiro)
	// ReturnDocument(options.After) retorna a struct já atualizada com o novo status
	opts := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "criado_em", Value: 1}}).
		SetReturnDocument(options.After)

	var letra model.Letra
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&letra)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoPendingLyrics
		}
		return nil, err
	}

	return &letra, nil
}

// AtualizarStatusMusica atualiza a letra após processamento (encontrada ou não).
func (r *LetraRepository) AtualizarStatusMusica(ctx context.Context, id interface{}, status model.StatusLetra, conteudo string, sincronizada bool) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":        status,
			"conteudo":      conteudo,
			"sincronizada":  sincronizada,
			"atualizado_em": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
