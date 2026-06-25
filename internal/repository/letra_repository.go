package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

// BuscarMusicaPendente busca a música PENDENTE mais antiga e altera atômicamente seu status para PROCESSANDO.
// Garante ausência de Race Conditions pelo uso de FindOneAndUpdate.
func (r *LetraRepository) BuscarMusicaPendente(ctx context.Context) (*model.Letra, error) {
	filter := bson.M{
		"status":     model.StatusPendente,
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
func (r *LetraRepository) AtualizarStatusMusica(ctx context.Context, id interface{}, status model.StatusLetra, conteudo string, sincronizada bool, fonte string) error {
	filter := bson.M{"_id": id}
	
	setFields := bson.M{
		"status":        status,
		"texto_letra":   conteudo,
		"sincronizada":  sincronizada,
		"atualizado_em": time.Now(),
	}
	
	if fonte != "" {
		setFields["fonte_letra"] = fonte
	} else {
		setFields["fonte_letra"] = nil
	}

	update := bson.M{
		"$set": setFields,
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
